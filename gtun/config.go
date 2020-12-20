package gtun

import (
	"io/ioutil"
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	ClientConfig *ClientConfig `toml:"client"`
}

func ParseConfig(path string) (*Config, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadAll(fp)
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
