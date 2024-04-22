package main

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	GeoConfig  GeoConfig                  `yaml:"Geo"`
	Fetcher    map[string]json.RawMessage `yaml:"fetcher"`
	GtunConfig *GtunConfig                `yaml:"gtun"`
	Log        Log                        `yaml:"log"`
}

type GeoConfig struct {
	GeoIPFile     string `yaml:"geo_ip_file"`
	GeoDomainFile string `yaml:"geo_domain_file"`
}

type GtunConfig struct {
	FetcherName string   `yaml:"fetcher_name"`
	RouteFile   string   `yaml:"route_file"`
	ProxyFile   string   `yaml:"proxy_file"`
	RestartCmd  []string `yaml:"restart_cmd"`
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
	return &conf, err
}
