package subsonic

import (
	"net/http"
	"strconv"

	"github.com/deluan/navidrome/engine"
	"github.com/deluan/navidrome/server/subsonic/responses"
	"github.com/deluan/navidrome/utils"
)

type StreamController struct {
	streamer engine.MediaStreamer
}

func NewStreamController(streamer engine.MediaStreamer) *StreamController {
	return &StreamController{streamer: streamer}
}

func (c *StreamController) Stream(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	id, err := RequiredParamString(r, "id", "id parameter required")
	if err != nil {
		return nil, err
	}
	maxBitRate := utils.ParamInt(r, "maxBitRate", 0)
	format := utils.ParamString(r, "format")

	ms, err := c.streamer.NewStream(r.Context(), id, maxBitRate, format)
	if err != nil {
		return nil, err
	}

	// Override Content-Type detected by http.FileServer
	w.Header().Set("Content-Type", ms.ContentType())
	w.Header().Set("X-Content-Duration", strconv.Itoa(ms.Duration()))
	http.ServeContent(w, r, ms.Name(), ms.ModTime(), ms)
	return nil, nil
}

func (c *StreamController) Download(w http.ResponseWriter, r *http.Request) (*responses.Subsonic, error) {
	id, err := RequiredParamString(r, "id", "id parameter required")
	if err != nil {
		return nil, err
	}

	ms, err := c.streamer.NewStream(r.Context(), id, 0, "raw")
	if err != nil {
		return nil, err
	}

	// Override Content-Type detected by http.FileServer
	w.Header().Set("Content-Type", ms.ContentType())
	http.ServeContent(w, r, ms.Name(), ms.ModTime(), ms)
	return nil, nil
}
