package gtun

import (
	"gopkg.in/yaml.v2"
	"os"
)

var gConfig *Config

type Config struct {
	Settings   map[string]RegionConfig `yaml:"settings"`
	HTTPServer HTTPConfig              `yaml:"http_server"`
	Log        Log                     `yaml:"log"`
}

type HTTPConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}

type RegionConfig struct {
	Route     []RouteConfig     `yaml:"route"`
	ProxyFile string            `yaml:"proxy_file"`
	Proxy     map[string]string `yaml:"proxy"`
}

type RouteConfig struct {
	Region    string `yaml:"region"`
	TraceAddr string `yaml:"trace_addr"`
	Scheme    string `yaml:"scheme"`
	Addr      string `yaml:"addr"`
	AuthKey   string `yaml:"auth_key"`
}

type Log struct {
	Days  int64  `yaml:"days"`
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

func ParseConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return ParseBuffer(content)
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

func GetConfig() *Config {
	return gConfig
}
