package network

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"IrChad/internal/chadd"
)

type Network struct {
	DiscoveryURL string
	Config       *Config
	chaddClient  *chadd.Client
	AccountName  string
	Nick         string
}

func NewNetwork(discoveryURL string, config *Config) *Network {
	return &Network{
		DiscoveryURL: discoveryURL,
		Config:       config,
		chaddClient:  chadd.NewChaddClient(discoveryURL),
	}
}

func (n *Network) Login(accountName, passphrase string) error {
	err := n.chaddClient.Login(accountName, passphrase)
	if err != nil {
		return err
	}

	n.AccountName = accountName
	return nil
}

func (n *Network) Identity() string {
	if n.AccountName != "" {
		return n.AccountName
	}
	return n.Nick
}

func (n *Network) GetJoinToken(room string) (string, error) {
	return n.chaddClient.GetJoinToken(n.Identity(), room)
}

func Discover(uri string) (*Config, error) {
	u, err := url.Parse(uri)
	u.Path = "/config.json"
	if err != nil {
		return nil, fmt.Errorf("invalid uri: %s", err.Error())
	}

	response, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("server config not found")
	}

	dec := json.NewDecoder(response.Body)
	defer response.Body.Close()

	var config Config
	err = dec.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
