package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/livekit/protocol/auth"
)

func getToken(identity, room string) (string, error) {
	key := os.Getenv("CHADD_LIVEKIT_KEY")
	secret := os.Getenv("CHADD_LIVEKIT_SECRET")

	at := auth.NewAccessToken(key, secret)
	videoGrant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	videoGrant.SetCanPublish(true)
	videoGrant.SetCanSubscribe(true)

	sipGrant := &auth.SIPGrant{
		Admin: false,
		Call:  true,
	}

	at.SetSIPGrant(sipGrant).SetVideoGrant(videoGrant).SetIdentity(identity).SetValidFor(time.Hour)
	return at.ToJWT()
}

type GetJoinTokenParams struct {
	Identity string `json:"identity"`
	Room     string `json:"room"`
}

type GetJoinTokenResponse struct {
	Token string `json:"token"`
}

func getJoinToken(w http.ResponseWriter, r *http.Request) {
	account, ok := GetAccount(r)
	if !ok && os.Getenv("CHADD_ALLOW_ANONYMOUS_LIVE") != "1" {
		http.Error(
			w,
			"you must be logged in to join",
			http.StatusForbidden,
		)
		return
	}

	var reqBody GetJoinTokenParams
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&reqBody)
	if err != nil {
		http.Error(
			w,
			"bad request",
			http.StatusBadRequest,
		)
		return
	}

	var callerIdentity string
	if account == "" {
		callerIdentity = reqBody.Identity
	} else {
		callerIdentity = account
	}

	t, err := getToken(callerIdentity, reqBody.Room)
	if err != nil {
		log.Printf("error creating join token: %s", err.Error())
		http.Error(
			w,
			"error creating join token",
			http.StatusInternalServerError,
		)
		return
	}

	enc := json.NewEncoder(w)
	_ = enc.Encode(&GetJoinTokenResponse{Token: t})
}
