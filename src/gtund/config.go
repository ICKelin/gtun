package main

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var gConfig *Config

type Config struct {
	EnableAuth   bool           `yaml:"enable_auth"`
	Auths        []AuthConfig   `yaml:"auths"`
	Trace        string         `yaml:"trace"`
	ServerConfig []ServerConfig `yaml:"server"`
	Log          Log            `yaml:"log"`
}

type Log struct {
	Days  int64  `yaml:"days"`
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

type AuthConfig struct {
	AccessToken string `yaml:"access_token"`
	ExpiredAt   int64  `yaml:"expired_at"`
}

func ParseConfig(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseConfig(content)
}

func parseConfig(content []byte) (*Config, error) {
	var c Config
	err := yaml.Unmarshal(content, &c)
	if err != nil {
		return nil, err
	}
	gConfig = &c
	return &c, err
}

func (c *Config) String() string {
	cnt, _ := json.MarshalIndent(c, "", "\t")
	return string(cnt)
}

func GetConfig() *Config {
	return gConfig
}
