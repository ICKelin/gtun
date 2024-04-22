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
	RouteFile string `yaml:"route_file"`
	ProxyFile string `yaml:"proxy_file"`
	Log       Log    `yaml:"log"`
}

type RouteConfig struct {
	Region  string `yaml:"region" json:"region"`
	Scheme  string `yaml:"scheme" json:"scheme"`
	Server  string `yaml:"server" json:"server"`
	Trace   string `yaml:"trace" json:"trace"`
	AuthKey string `yaml:"auth_key" json:"auth_key"`
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

func ParseRoute(routeFile string) ([]*RouteConfig, error) {
	content, err := os.ReadFile(routeFile)
	if err != nil {
		return nil, err
	}

	configContent, err := signature.UnSign(content)
	if err != nil {
		return nil, err
	}

	var routeConfig = make([]*RouteConfig, 0)
	err = json.Unmarshal(configContent, &routeConfig)
	if err != nil {
		return nil, err
	}

	return routeConfig, nil
}
