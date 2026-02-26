package live

import (
	"io"
	"net/http"
)

type VideoTrackHandler struct {
	stream *VideoStream
}

func (h *VideoTrackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "multipart/x-mixed-replace; boundary=irchad")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("connection", "keep-alive")
	io.Copy(w, h.stream)
}

func (h *VideoTrackHandler) Close() {
	h.stream.Close()
}
