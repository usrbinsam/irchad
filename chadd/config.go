package main

import (
	"io"
	"net/http"
	"os"
)

func getConfig(w http.ResponseWriter, _ *http.Request) {
	f, err := os.Open("/config.json")
	if err != nil {
		http.Error(
			w, "/config.json not found", http.StatusInternalServerError,
		)
		return
	}
	w.Header().Add("content-type", "application/json")
	io.Copy(w, f)
}
