package live

import (
	"net/http"
)

type StreamHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Close()
}
