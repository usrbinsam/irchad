package network

import (
	"fmt"
	"log"
	"sync"
)

type NetworkService struct {
	mu       sync.RWMutex
	networks map[string]*Network
}

func NewNetworkService() *NetworkService {
	return &NetworkService{
		mu:       sync.RWMutex{},
		networks: make(map[string]*Network),
	}
}

func (s *NetworkService) Connect(discoveryURL, nick, accountName, passphrase string) (*Config, error) {
	config, err := Discover(discoveryURL)
	if err != nil {
		return nil, err
	}
	log.Println("ok?")

	n := NewNetwork(discoveryURL, config)
	n.AccountName = accountName
	n.Nick = nick
	if accountName != "" && passphrase != "" {
		log.Println("login?")
		err = n.Login(accountName, passphrase)
		log.Println("done login?")
		if err != nil {
			log.Fatalf("failed to login: %s", err.Error())
		}
	}
	s.mu.Lock()
	s.networks[discoveryURL] = n
	s.mu.Unlock()

	return config, nil
}

func (s *NetworkService) Get(discoveryURL string) *Network {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.networks[discoveryURL]
	if !ok {
		return nil
	}
	return n
}

func (s *NetworkService) GetJoinToken(discoveryURL, channelName string) (string, error) {
	n := s.Get(discoveryURL)
	if n == nil {
		return "", fmt.Errorf("no such network %s", discoveryURL)
	}

	return n.GetJoinToken(channelName)
}
