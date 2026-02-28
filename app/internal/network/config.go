package network

type LiveConfig struct {
	Server string `json:"server"`
}

type IRCConfig struct {
	Server string `json:"server"`
}

type Config struct {
	Name string     `json:"name"`
	IRC  IRCConfig  `json:"irc"`
	Live LiveConfig `json:"live"`
	Auth string     `json:"auth"`
}
