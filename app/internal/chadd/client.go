package chadd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Client struct {
	baseURL   string
	client    *http.Client
	authToken string
}

func NewChaddClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type LoginRequest struct {
	AccountName string `json:"accountName"`
	Passphrase  string `json:"passphrase"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (c *Client) requestFactory(method, path string, body io.Reader) (*http.Request, error) {
	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if c.authToken != "" {
		req.Header.Add("authorization", "Bearer "+c.authToken)
	}

	return req, nil
}

func (c *Client) Login(accountName, passphrase string) error {
	log.Println("chadd login?")
	body := bytes.Buffer{}
	enc := json.NewEncoder(&body)
	_ = enc.Encode(&LoginRequest{AccountName: accountName, Passphrase: passphrase})
	req, err := c.requestFactory("POST", "/api/login", &body)
	if err != nil {
		return err
	}

	log.Println("Do?")
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	log.Println("done?")
	if res.StatusCode != 200 {
		return fmt.Errorf("non-200 from chadd: %d", res.StatusCode)
	}

	var resBody LoginResponse
	log.Println("decode?")
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&resBody)
	if err != nil {
		return err
	}

	c.authToken = resBody.Token

	return nil
}

type JoinTokenResponse struct {
	Token string `json:"token"`
}

func (c *Client) GetJoinToken(identity, room string) (string, error) {
	body := bytes.Buffer{}
	enc := json.NewEncoder(&body)
	_ = enc.Encode(map[string]string{"identity": identity, "room": room})

	req, err := c.requestFactory("POST", "/api/join", &body)
	if err != nil {
		return "", err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("non-200 from chadd: %d", res.StatusCode)
	}

	var resBody JoinTokenResponse
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&resBody)
	if err != nil {
		return "", err
	}

	return resBody.Token, nil
}
