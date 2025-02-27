package conf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deluan/navidrome/consts"
	"github.com/deluan/navidrome/log"
	"github.com/koding/multiconfig"
)

type nd struct {
	Port        string `default:"4533"`
	MusicFolder string `default:"./music"`
	DataFolder  string `default:"./"`
	DbPath      string
	LogLevel    string `default:"info"`

	IgnoredArticles string `default:"The El La Los Las Le Les Os As O A"`
	IndexGroups     string `default:"A B C D E F G H I J K L M N O P Q R S T U V W X-Z(XYZ) [Unknown]([)"`

	EnableDownsampling bool   `default:"false"`
	MaxBitRate         int    `default:"0"`
	DownsampleCommand  string `default:"ffmpeg -i %s -map 0:0 -b:a %bk -v 0 -f mp3 -"`
	ProbeCommand       string `default:"ffmpeg -i %s -f ffmetadata"`
	ScanInterval       string `default:"1m"`

	// DevFlags. These are used to enable/disable debugging and incomplete features
	DevDisableBanner           bool   `default:"false"`
	DevLogSourceLine           bool   `default:"false"`
	DevAutoCreateAdminPassword string `default:""`
}

var Server = &nd{}

func newWithPath(path string, skipFlags ...bool) *multiconfig.DefaultLoader {
	var loaders []multiconfig.Loader

	// Read default values defined via tag fields "default"
	loaders = append(loaders, &multiconfig.TagLoader{})

	if _, err := os.Stat(path); err == nil {
		if strings.HasSuffix(path, "toml") {
			loaders = append(loaders, &multiconfig.TOMLLoader{Path: path})
		}

		if strings.HasSuffix(path, "json") {
			loaders = append(loaders, &multiconfig.JSONLoader{Path: path})
		}

		if strings.HasSuffix(path, "yml") || strings.HasSuffix(path, "yaml") {
			loaders = append(loaders, &multiconfig.YAMLLoader{Path: path})
		}
	}

	e := &multiconfig.EnvironmentLoader{}
	loaders = append(loaders, e)
	if len(skipFlags) == 0 || !skipFlags[0] {
		f := &multiconfig.FlagLoader{}
		loaders = append(loaders, f)
	}

	loader := multiconfig.MultiLoader(loaders...)

	d := &multiconfig.DefaultLoader{}
	d.Loader = loader
	d.Validator = multiconfig.MultiValidator(&multiconfig.RequiredValidator{})
	return d
}

func LoadFromFile(confFile string, skipFlags ...bool) {
	m := newWithPath(confFile, skipFlags...)
	err := m.Load(Server)
	if err == flag.ErrHelp {
		os.Exit(1)
	}
	if err != nil {
		fmt.Printf("Error trying to load config '%s'. Error: %v", confFile, err)
		os.Exit(2)
	}
	if Server.DbPath == "" {
		Server.DbPath = filepath.Join(Server.DataFolder, "navidrome.db")
	}
	if os.Getenv("PORT") != "" {
		Server.Port = os.Getenv("PORT")
	}
	log.SetLevelString(Server.LogLevel)
	log.SetLogSourceLine(Server.DevLogSourceLine)
	log.Trace("Loaded configuration", "file", confFile, "config", fmt.Sprintf("%#v", Server))
}

func Load() {
	LoadFromFile(consts.LocalConfigFile)
}
