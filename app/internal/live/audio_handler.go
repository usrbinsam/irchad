package live

import (
	"log"
	"net/http"
)

type AudioTrackHandler struct {
	stream *OpusStream
}

func (h *AudioTrackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("channel closed")
			return
		}
		_, err := w.Write(b)
		// log.Printf("wrote %d bytes to browser", n)
		flusher.Flush()
		if err != nil {
			// log.Printf("write error: %s", err.Error())
			// close(ch)
			break
		}

	}

	log.Printf("stream closed")
	// log.Printf("streaming audio for %s", participantID)
	// n, _ := io.Copy(w, stream)
	// log.Printf("stream closed. sent %d MB\n", n/1024)
}

func (h *AudioTrackHandler) Close() {
}
