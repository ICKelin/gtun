package gtun

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	ClientConfig *ClientConfig `toml:"client"`
	Log          Log           `toml:"log"`
}

type Log struct {
	Days  int64  `toml:"days"`
	Level string `toml:"level"`
	Path  string `toml:"path"`
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
