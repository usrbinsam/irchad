package live

import (
	"context"
	"log"
	"net"
	"net/http"
)

func newHTTPServer(mux *http.ServeMux) *httpServer {
	return &httpServer{mux: mux}
}

// httpServer provides a simple way to launch a
// HTTP listener on any available port on the system,
// and then shut it down as neccessary
type httpServer struct {
	listener net.Listener
	server   *http.Server
	mux      *http.ServeMux
}

func (d *httpServer) Start() error {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}

	d.listener = l
	d.server = &http.Server{Handler: d.mux}
	go func() {
		_ = d.server.Serve(l)
	}()

	return nil
}

func (d *httpServer) Shutdown(ctx context.Context) error {
	return d.server.Shutdown(ctx)
}

func (d *httpServer) Addr() string {
	return d.listener.Addr().String()
}

func (l *LiveChat) startDecodeServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/stream", l.serveStream)
	l.decoderServer = newHTTPServer(mux)
	err := l.decoderServer.Start()
	if err != nil {
		log.Printf("decode server exited: %s", err.Error())
		return err
	}

	log.Printf("decode server listening on %s\n", l.decoderServer.Addr())
	return err
}

func (l *LiveChat) serveStream(w http.ResponseWriter, r *http.Request) {
	qstring := r.URL.Query()
	participantID := qstring.Get("pid")
	trackID := qstring.Get("tid")

	l.registry.mu.RLock()
	tracks, ok := l.registry.streams[participantID]

	var handler StreamHandler

	if ok {
		handler = tracks[trackID]
	}

	if handler == nil {
		http.NotFound(w, r)
		return
	}

	handler.ServeHTTP(w, r)
}
