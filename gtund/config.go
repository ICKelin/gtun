package gtund

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Name            string           `toml:"name"`
	ServerConfig    *ServerConfig    `toml:"server"`
	DHCPConfig      *DHCPConfig      `toml:"dhcp"`
	InterfaceConfig *InterfaceConfig `toml:"interface"`
	ReverseConfig   *ReverseConfig   `toml:"reverse"`
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
	return &c, err
}

func (c *Config) String() string {
	cnt, _ := json.MarshalIndent(c, "", "\t")
	return string(cnt)
}
