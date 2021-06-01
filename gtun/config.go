package gtun

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Forwards map[string]ForwardConfig `yaml:"forwards"`
	Log      Log                      `yaml:"log"`
}

type ForwardConfig struct {
	ServerAddr string           `yaml:"server"`
	AuthKey    string           `yaml:"authKey"`
	TCPForward TCPForwardConfig `yaml:"tcp"`
	UDPForward UDPForwardConfig `yaml:"udp"`
	Transport  TransportConfig  `yaml:"transport"`
}

type TCPForwardConfig struct {
	ListenAddr   string `yaml:"listen"`
	ReadTimeout  int    `yaml:"readTimeout"`
	WriteTimeout int    `yaml:"writeTimeout"`
}

type UDPForwardConfig struct {
	ListenAddr     string `yaml:"listen"`
	ReadTimeout    int    `yaml:"readTimeout"`
	WriteTimeout   int    `yaml:"writeTimeout"`
	SessionTimeout int    `yaml:"sessionTimeout"`
}

type TransportConfig struct {
	Scheme        string `yaml:"scheme"`
	ConfigContent string `yaml:"config"`
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
