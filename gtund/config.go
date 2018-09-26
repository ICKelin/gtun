package gtund

import (
	"encoding/json"
	"io/ioutil"
)

var config Config

type Config struct {
	Region string     `json:"region"`
	GodCfg *GodConfig `json:"god_config"`
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
	err := json.Unmarshal(content, &c)
	config = c
	return &c, err
}

func (c *Config) String() string {
	cnt, _ := json.MarshalIndent(c, "", "\t")
	return string(cnt)
}

func GetConfig() *Config {
	return &config
}
