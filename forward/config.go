package forward

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	ListenerConfig ListenerConfig `yaml:"listener"`
	NexthopConfig  NextHopConfig  `yaml:"dialer"`
}

type ListenerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
	Scheme     string `yaml:"scheme"`
	RawConfig  string `yaml:"raw_config"`
}

type NextHopConfig struct {
	NexthopAddr string `yaml:"nexthop_addr"`
	Scheme      string `yaml:"scheme"`
	RawConfig   string `yaml:"raw_config"`
}

func ParseConfig(path string) (*Config, error) {
	cnt, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg = Config{}
	err = yaml.Unmarshal(cnt, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, err
}
