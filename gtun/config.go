package gtun

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	ClientConfig *ClientConfig    `toml:"client"`
	TCPForward   TCPForwardConfig `toml:"tcpforward"`
	UDPForward   UDPForwardConfig `toml:"udpforward"`
	Log          Log              `toml:"log"`
}

type Log struct {
	Days  int64  `toml:"days"`
	Level string `toml:"level"`
	Path  string `toml:"path"`
}

type TCPForwardConfig struct {
	ListenAddr   string `toml:"listen"`
	ReadTimeout  int    `toml:"readTimeout"`
	WriteTimeout int    `toml:"writeTimeout"`
}

type UDPForwardConfig struct {
	ListenAddr     string `toml:"listen"`
	ReadTimeout    int    `toml:"readTimeout"`
	WriteTimeout   int    `toml:"writeTimeout"`
	SessionTimeout int    `toml:"sessionTimeout"`
}

func ParseConfig(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseConfig(content)
}

func parseConfig(content []byte) (*Config, error) {
	conf := Config{}
	err := toml.Unmarshal(content, &conf)
	return &conf, err
}
