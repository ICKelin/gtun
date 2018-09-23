package gtund

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
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
	return &c, err
}

func (c *Config) String() string {
	cnt, _ := json.MarshalIndent(c, "", "\t")
	return string(cnt)
}
