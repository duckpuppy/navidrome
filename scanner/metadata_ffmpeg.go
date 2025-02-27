package scanner

import (
	"bufio"
	"errors"
	"mime"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deluan/navidrome/conf"
	"github.com/deluan/navidrome/log"
)

type Metadata struct {
	filePath string
	suffix   string
	fileInfo os.FileInfo
	tags     map[string]string
}

func (m *Metadata) Title() string               { return m.tags["title"] }
func (m *Metadata) Album() string               { return m.tags["album"] }
func (m *Metadata) Artist() string              { return m.tags["artist"] }
func (m *Metadata) AlbumArtist() string         { return m.tags["album_artist"] }
func (m *Metadata) Composer() string            { return m.tags["composer"] }
func (m *Metadata) Genre() string               { return m.tags["genre"] }
func (m *Metadata) Year() int                   { return m.parseYear("year") }
func (m *Metadata) TrackNumber() (int, int)     { return m.parseTuple("trackNum", "trackTotal") }
func (m *Metadata) DiscNumber() (int, int)      { return m.parseTuple("discNum", "discTotal") }
func (m *Metadata) HasPicture() bool            { return m.tags["hasPicture"] == "Video" }
func (m *Metadata) Comment() string             { return m.tags["comment"] }
func (m *Metadata) Compilation() bool           { return m.parseBool("compilation") }
func (m *Metadata) Duration() int               { return m.parseDuration("duration") }
func (m *Metadata) BitRate() int                { return m.parseInt("bitrate") }
func (m *Metadata) ModificationTime() time.Time { return m.fileInfo.ModTime() }
func (m *Metadata) FilePath() string            { return m.filePath }
func (m *Metadata) Suffix() string              { return m.suffix }
func (m *Metadata) Size() int                   { return int(m.fileInfo.Size()) }

func LoadAllAudioFiles(dirPath string) (map[string]os.FileInfo, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}
	audioFiles := make(map[string]os.FileInfo)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		filePath := filepath.Join(dirPath, f.Name())
		extension := path.Ext(filePath)
		if !isAudioFile(extension) {
			continue
		}
		fi, err := os.Stat(filePath)
		if err != nil {
			log.Error("Could not stat file", "filePath", filePath, err)
		} else {
			audioFiles[filePath] = fi
		}
	}

	return audioFiles, nil
}

func ExtractAllMetadata(inputs []string) (map[string]*Metadata, error) {
	cmdLine, args := createProbeCommand(inputs)

	log.Trace("Executing command", "arg0", cmdLine, "args", args)
	cmd := exec.Command(cmdLine, args...)
	output, _ := cmd.CombinedOutput()
	mds := map[string]*Metadata{}
	if len(output) == 0 {
		return mds, errors.New("error extracting metadata files")
	}
	infos := parseOutput(string(output))
	for file, info := range infos {
		md, err := extractMetadata(file, info)
		// Skip files with errors
		if err == nil {
			mds[file] = md
		}
	}
	return mds, nil
}

var inputRegex = regexp.MustCompile(`(?m)^Input #\d+,.*,\sfrom\s'(.*)'`)

func parseOutput(output string) map[string]string {
	split := map[string]string{}
	all := inputRegex.FindAllStringSubmatchIndex(output, -1)
	for i, loc := range all {
		// Filename is the first captured group
		file := output[loc[2]:loc[3]]

		// File info is everything from the match, up until the beginning of the next match
		info := ""
		initial := loc[1]
		if i < len(all)-1 {
			end := all[i+1][0] - 1
			info = output[initial:end]
		} else {
			// if this is the last match
			info = output[initial:]
		}
		split[file] = info
	}
	return split
}

func extractMetadata(filePath, info string) (*Metadata, error) {
	m := &Metadata{filePath: filePath, tags: map[string]string{}}
	m.suffix = strings.ToLower(strings.TrimPrefix(path.Ext(filePath), "."))
	var err error
	m.fileInfo, err = os.Stat(filePath)
	if err != nil {
		log.Warn("Error stating file. Skipping", "filePath", filePath, err)
		return nil, errors.New("error stating file")
	}

	m.parseInfo(info)
	if len(m.tags) == 0 {
		log.Trace("Not a media file. Skipping", "filePath", filePath)
		return nil, errors.New("not a media file")
	}
	return m, nil
}

func isAudioFile(extension string) bool {
	typ := mime.TypeByExtension(extension)
	return strings.HasPrefix(typ, "audio/")
}

var (
	tagsRx = map[*regexp.Regexp]string{
		regexp.MustCompile(`(?i)^\s{4,6}compilation\s+:(.*)`):    "compilation",
		regexp.MustCompile(`(?i)^\s{4,6}genre\s+:\s(.*)`):        "genre",
		regexp.MustCompile(`(?i)^\s{4,6}title\s+:\s(.*)`):        "title",
		regexp.MustCompile(`(?i)^\s{4,6}comment\s+:\s(.*)`):      "comment",
		regexp.MustCompile(`(?i)^\s{4,6}artist\s+:\s(.*)`):       "artist",
		regexp.MustCompile(`(?i)^\s{4,6}album_artist\s+:\s(.*)`): "album_artist",
		regexp.MustCompile(`(?i)^\s{4,6}TCM\s+:\s(.*)`):          "composer",
		regexp.MustCompile(`(?i)^\s{4,6}album\s+:\s(.*)`):        "album",
		regexp.MustCompile(`(?i)^\s{4,6}track\s+:\s(.*)`):        "trackNum",
		regexp.MustCompile(`(?i)^\s{4,6}tracktotal\s+:\s(.*)`):   "trackTotal",
		regexp.MustCompile(`(?i)^\s{4,6}disc\s+:\s(.*)`):         "discNum",
		regexp.MustCompile(`(?i)^\s{4,6}disctotal\s+:\s(.*)`):    "discTotal",
		regexp.MustCompile(`(?i)^\s{4,6}TPA\s+:\s(.*)`):          "discNum",
		regexp.MustCompile(`(?i)^\s{4,6}date\s+:\s(.*)`):         "year",
		regexp.MustCompile(`^\s{4}Stream #\d+:\d+: (Video):.*`):  "hasPicture",
	}

	durationRx = regexp.MustCompile(`^\s\sDuration: ([\d.:]+).*bitrate: (\d+)`)
)

func (m *Metadata) parseInfo(info string) {
	reader := strings.NewReader(info)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		for rx, tag := range tagsRx {
			// Skip when the tag was previously found
			if _, ok := m.tags[tag]; ok {
				continue
			}
			match := rx.FindStringSubmatch(line)
			if len(match) > 0 {
				m.tags[tag] = match[1]
				break
			}
			match = durationRx.FindStringSubmatch(line)
			if len(match) == 0 {
				continue
			}
			m.tags["duration"] = match[1]
			if len(match) > 1 {
				m.tags["bitrate"] = match[2]
			}
		}
	}
}

func (m *Metadata) parseInt(tagName string) int {
	if v, ok := m.tags[tagName]; ok {
		i, _ := strconv.Atoi(v)
		return i
	}
	return 0
}

var tagYearFormats = []string{
	"2006",
	"2006.01",
	"2006.01.02",
	"2006-01",
	"2006-01-02",
	time.RFC3339,
}

var dateRegex = regexp.MustCompile(`^([12]\d\d\d)`)

func (m *Metadata) parseYear(tagName string) int {
	if v, ok := m.tags[tagName]; ok {
		match := dateRegex.FindStringSubmatch(v)
		if len(match) == 0 {
			log.Error("Error parsing year from ffmpeg date field. Please report this issue", "file", m.filePath, "date", v)
			return 0
		}
		year, _ := strconv.Atoi(match[1])
		return year
	}
	return 0
}

func (m *Metadata) parseTuple(numTag string, totalTag string) (int, int) {
	if v, ok := m.tags[numTag]; ok {
		tuple := strings.Split(v, "/")
		t1, t2 := 0, 0
		t1, _ = strconv.Atoi(tuple[0])
		if len(tuple) > 1 {
			t2, _ = strconv.Atoi(tuple[1])
		} else {
			t2, _ = strconv.Atoi(m.tags[totalTag])
		}
		return t1, t2
	}
	return 0, 0
}

func (m *Metadata) parseBool(tagName string) bool {
	if v, ok := m.tags[tagName]; ok {
		i, _ := strconv.Atoi(strings.TrimSpace(v))
		return i == 1
	}
	return false
}

var zeroTime = time.Date(0000, time.January, 1, 0, 0, 0, 0, time.UTC)

func (m *Metadata) parseDuration(tagName string) int {
	if v, ok := m.tags[tagName]; ok {
		d, err := time.Parse("15:04:05", v)
		if err != nil {
			return 0
		}
		return int(d.Sub(zeroTime).Seconds())
	}
	return 0
}

func createProbeCommand(inputs []string) (string, []string) {
	cmd := conf.Server.ProbeCommand

	split := strings.Split(cmd, " ")
	args := make([]string, 0)
	first := true
	for _, s := range split {
		if s == "%s" {
			for _, inp := range inputs {
				if !first {
					args = append(args, "-i")
				}
				args = append(args, inp)
				first = false
			}
			continue
		}
		args = append(args, s)
	}

	return args[0], args[1:]
}
