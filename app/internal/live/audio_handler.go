package live

import (
	"log"
	"net/http"
)

type AudioTrackHandler struct {
	stream *OpusStream
}

func (h *AudioTrackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}
	w.Header().Set("content-type", "audio/ogg")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("connection", "keep-alive")

	flusher, canFlush := w.(http.Flusher)
	if !canFlush {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	opusHeader, ch := h.stream.Subscribe()
	if _, err := w.Write(opusHeader); err != nil {
		log.Printf("writing opus header failed: %s", err.Error())
		return
	}

	for {
		b, ok := <-ch
		if !ok {
			log.Printf("channel closed for %+v", h)
			return
		}
		_, err := w.Write(b)
		flusher.Flush()
		if err != nil {
			break
		}

	}

	log.Printf("stream closed")
}

func (h *AudioTrackHandler) Close() {
}
