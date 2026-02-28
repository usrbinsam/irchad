package main

import (
	"log"
	"net/http"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		log.Printf(
			"method=%s path=%s ip=%s",
			r.Method, r.URL.Path, r.RemoteAddr,
		)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, err := Verify(w, r)
		if err != nil {
			log.Printf("verification error: %s", err.Error())
			next.ServeHTTP(w, r)
		} else {
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
