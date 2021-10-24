package gtund

import (
	"encoding/json"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerConfig   ServerConfig `yaml:"server"`
	Scheme         string       `yaml:"scheme"`
	ListenerConfig string       `yaml:"listenerConfig"`
	Log            Log          `yaml:"log"`
}

type Log struct {
	Days  int64  `yaml:"days"`
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
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
	return &c, err
}

func (c *Config) String() string {
	cnt, _ := json.MarshalIndent(c, "", "\t")
	return string(cnt)
}
