package main

import (
	"log"
	"net/http"
	"os"
)

var ergoClient *ErgoClient

func main() {
	ergoClient = &ErgoClient{
		bearerToken: os.Getenv("CHADD_ERGO_BEARER_TOKEN"),
		baseURL:     os.Getenv("CHADD_ERGO_BASE_URL"),
		client:      &http.Client{},
	}

	http.Handle("POST /api/login", LoggerMiddleware(http.HandlerFunc(login)))
	http.Handle("POST /api/join", LoggerMiddleware(AuthMiddleware(http.HandlerFunc(getJoinToken))))
	http.Handle("GET /config.json", LoggerMiddleware(http.HandlerFunc(getConfig)))

	log.Printf("chadd running")
	err := http.ListenAndServe("0.0.0.0:8888", nil)
	if err != nil {
		log.Println(err.Error())
	}
}
