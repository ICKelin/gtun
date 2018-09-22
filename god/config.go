package god

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Listener    string       `json:"listener"`
	GtundConfig *gtundConfig `json:"gtund_config"`
	GtunConfig  *gtunConfig  `json:"gtun_config"`
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
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) String() string {
	bytes, _ := json.MarshalIndent(c, "", "\t")
	return string(bytes)
}
