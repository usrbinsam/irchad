package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type ErgoClient struct {
	client      *http.Client
	baseURL     string
	bearerToken string
}

func (e *ErgoClient) requestFactory(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	url := e.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("authorization", "Bearer "+e.bearerToken)
	return req, nil
}

type CheckAuthParams struct {
	AccountName string `json:"accountName"`
	Passphrase  string `json:"passphrase"`
}

type CheckAuthResponse struct {
	Success     bool   `json:"success"`
	AccountName string `json:"accountName"`
}

func (e *ErgoClient) CheckAuth(ctx context.Context, body *CheckAuthParams) (*CheckAuthResponse, error) {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	err := enc.Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := e.requestFactory(ctx, "POST", "/v1/account_details", &buf)
	if err != nil {
		return nil, err
	}

	res, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("non-200 from ergo API: %d", res.StatusCode)
		return nil, fmt.Errorf("unable to check at this time, try again later")
	}

	var resBody CheckAuthResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&resBody)
	if err != nil {
		log.Printf("invalid response body from ergo: %s", err.Error())
		return nil, fmt.Errorf("unable to check at this time, try again later")
	}

	return &resBody, nil
}
