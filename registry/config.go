package registry

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	GtundConfig *GtundConfig `toml:"gtund"`
	GtunConfig  *GtunConfig  `toml:"gtun"`
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
	err := toml.Unmarshal(content, &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) String() string {
	bytes, _ := json.MarshalIndent(c, "", "\t")
	return string(bytes)
}
