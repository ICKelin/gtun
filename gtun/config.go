package gtun

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Settings map[string]RegionConfig `yaml:"settings"`
	Log      Log                     `yaml:"log"`
}

type RegionConfig struct {
	Route []RouteConfig     `yaml:"route"`
	Proxy map[string]string `yaml:"proxy"`
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
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseConfig(content)
}

func parseConfig(content []byte) (*Config, error) {
	conf := Config{}
	err := yaml.Unmarshal(content, &conf)
	return &conf, err
}
