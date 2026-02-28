package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func signLoginToken(secret []byte, nick string) (string, error) {
	now := time.Now()
	exp := now.Add(24 * time.Hour)
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": nick,
			"iat": now.Unix(),
			"exp": exp.Unix(),
		},
	)
	return token.SignedString(secret)
}

func login(w http.ResponseWriter, r *http.Request) {
	var reqBody CheckAuthParams
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&reqBody)
	if err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	res, err := ergoClient.CheckAuth(r.Context(), &reqBody)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusInternalServerError,
		)
		return
	}

	if !res.Success {
		http.Error(
			w,
			"invalid username or password",
			http.StatusUnauthorized,
		)
		return
	}

	secret := []byte(os.Getenv("CHADD_AUTH_SECRET") + reqBody.Passphrase)
	signed, err := signLoginToken(secret, reqBody.AccountName)
	if err != nil {
		log.Printf("error signing token: %s", err.Error())
		http.Error(
			w,
			"cannot process your request right now",
			http.StatusInternalServerError,
		)
		return
	}
	accessToken := &http.Cookie{
		Name:     "access-token",
		Value:    signed,
		Domain:   os.Getenv("CHADD_AUTH_DOMAIN"),
		Path:     "/api",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, accessToken)
}
