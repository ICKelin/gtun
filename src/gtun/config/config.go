package config

import (
	"encoding/json"
	"github.com/ICKelin/gtun/src/internal/signature"
	"gopkg.in/yaml.v2"
	"os"
)

var gConfig *Config
var signatureKey = os.Getenv("GTUN_SIGNATURE")

type Config struct {
	Accelerator map[string]Accelerator `yaml:"accelerator"`
	Log         Log                    `yaml:"log"`
}

type RouteConfig struct {
	Scheme  string `yaml:"scheme" json:"scheme"`
	Server  string `yaml:"server" json:"server"`
	Trace   string `yaml:"trace" json:"trace"`
	AuthKey string `yaml:"auth_key" json:"auth_key"`
}

type Accelerator struct {
	Region  string                     `json:"region"`
	GeoSite []string                   `json:"geo_site"`
	GeoIP   []string                   `json:"geo_ip"`
	Routes  []*RouteConfig             `json:"routes"`
	Proxy   map[string]json.RawMessage `json:"proxy"`
}

type Log struct {
	Days  int64  `yaml:"days"`
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

func Parse(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configContent, err := signature.UnSign(content)
	if err != nil {
		return nil, err
	}

	return ParseBuffer(configContent)
}

func ParseBuffer(content []byte) (*Config, error) {
	conf := Config{}
	err := yaml.Unmarshal(content, &conf)
	if err != nil {
		return nil, err
	}
	gConfig = &conf
	return &conf, err
}

func ParseProxy(proxyFile string) (map[string]map[string]string, error) {
	content, err := os.ReadFile(proxyFile)
	if err != nil {
		return nil, err
	}

	configContent, err := signature.UnSign(content)
	if err != nil {
		return nil, err
	}

	proxies := make(map[string]map[string]string)
	err = yaml.Unmarshal(configContent, &proxies)
	if err != nil {
		return nil, err
	}
	return proxies, nil
}
