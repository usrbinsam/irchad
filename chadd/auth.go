package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
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

type LoginResponse struct {
	Token string `json:"token"`
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
		log.Printf("err from ergo: %s", err.Error())
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

	secret := []byte(os.Getenv("CHADD_AUTH_SECRET"))
	// we should always sign the casefolded version of the account name
	signed, err := signLoginToken(secret, res.AccountName)
	if err != nil {
		log.Printf("error signing token: %s", err.Error())
		http.Error(
			w,
			"cannot process your request right now",
			http.StatusInternalServerError,
		)
		return
	}

	enc := json.NewEncoder(w)
	_ = enc.Encode(&LoginResponse{Token: signed})
}

type AccountKey string

var Account AccountKey = "Account"

func Verify(w http.ResponseWriter, r *http.Request) (context.Context, error) {
	authorization := r.Header.Get("authorization")
	if authorization == "" {
		return nil, fmt.Errorf("no authorization header")
	}

	fields := strings.Fields(authorization)
	if len(fields) != 2 {
		return nil, fmt.Errorf("malformed authorization header")
	}

	if strings.ToLower(fields[0]) != "bearer" {
		return nil, fmt.Errorf("expected bearer token, got: %s", fields[0])
	}

	token, err := jwt.Parse(
		fields[1],
		func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(os.Getenv("CHADD_AUTH_SECRET")), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("jwt error: %s", err.Error())
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token: %v", token)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	ctx := context.WithValue(r.Context(), Account, claims["sub"])
	return ctx, nil
}

func GetAccount(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(Account).(string)
	return v, ok
}
