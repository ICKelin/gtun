package gtund

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Name         string       `toml:"name"`   // instance name
	IsTap        bool         `toml:"istap"`  // is tap device(supported for windows)
	ServerConfig ServerConfig `toml:"server"` // tcp server configuration
	DHCPConfig   DHCPConfig   `toml:"dhcp"`   // ip block configuration
	Log          Log          `toml:"log"`
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
	var c Config
	err := toml.Unmarshal(content, &c)
	return &c, err
}

func (c *Config) String() string {
	cnt, _ := json.MarshalIndent(c, "", "\t")
	return string(cnt)
}
