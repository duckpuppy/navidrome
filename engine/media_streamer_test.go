package engine

import (
	"time"

	"github.com/deluan/navidrome/conf"
	"github.com/deluan/navidrome/log"
	"github.com/deluan/navidrome/model"
	"github.com/deluan/navidrome/persistence"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MediaStreamer", func() {

	var streamer MediaStreamer
	var ds model.DataStore
	ctx := log.NewContext(nil)

	BeforeEach(func() {
		conf.Server.EnableDownsampling = true
		ds = &persistence.MockDataStore{}
		ds.MediaFile(ctx).(*persistence.MockMediaFile).SetData(`[{"id": "123", "path": "tests/fixtures/test.mp3", "bitRate": 128}]`, 1)
		streamer = NewMediaStreamer(ds)
	})

	Context("NewStream", func() {
		It("returns a rawMediaStream if format is 'raw'", func() {
			Expect(streamer.NewStream(ctx, "123", 0, "raw")).To(BeAssignableToTypeOf(&rawMediaStream{}))
		})
		It("returns a rawMediaStream if maxBitRate is 0", func() {
			Expect(streamer.NewStream(ctx, "123", 0, "mp3")).To(BeAssignableToTypeOf(&rawMediaStream{}))
		})
		It("returns a rawMediaStream if maxBitRate is higher than file bitRate", func() {
			Expect(streamer.NewStream(ctx, "123", 256, "mp3")).To(BeAssignableToTypeOf(&rawMediaStream{}))
		})
		It("returns a transcodedMediaStream if maxBitRate is lower than file bitRate", func() {
			s, err := streamer.NewStream(ctx, "123", 64, "mp3")
			Expect(err).To(BeNil())
			Expect(s).To(BeAssignableToTypeOf(&transcodedMediaStream{}))
			Expect(s.(*transcodedMediaStream).bitRate).To(Equal(64))
		})
	})

	Context("rawMediaStream", func() {
		var rawStream mediaStream
		var modTime time.Time

		BeforeEach(func() {
			modTime = time.Now()
			mf := &model.MediaFile{ID: "123", Path: "test.mp3", UpdatedAt: modTime, Suffix: "mp3"}
			rawStream = &rawMediaStream{mf: mf, ctx: ctx}
		})

		It("returns the ContentType", func() {
			Expect(rawStream.ContentType()).To(Equal("audio/mpeg"))
		})

		It("returns the ModTime", func() {
			Expect(rawStream.ModTime()).To(Equal(modTime))
		})
	})

	Context("createTranscodeCommand", func() {
		BeforeEach(func() {
			conf.Server.DownsampleCommand = "ffmpeg -i %s -b:a %bk mp3 -"
		})
		It("creates a valid command line", func() {
			cmd, args := createTranscodeCommand("/music library/file.mp3", 123, "")
			Expect(cmd).To(Equal("ffmpeg"))
			Expect(args).To(Equal([]string{"-i", "/music library/file.mp3", "-b:a", "123k", "mp3", "-"}))
		})

	})
})
