package gtun

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Forwards []ForwardConfig `yaml:"forwards"`
	Log      Log             `yaml:"log"`
}

type ForwardConfig struct {
	Region     string            `yaml:"region"`
	TCPForward TCPForwardConfig  `yaml:"tcp"`
	UDPForward UDPForwardConfig  `yaml:"udp"`
	Transport  []TransportConfig `yaml:"transport"`
	Ratelimit  uint64            `yaml:"rateLimit"` // rate limit mbps
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
	Server        string `yaml:"server"`
	AuthKey       string `yaml:"authKey"`
	Scheme        string `yaml:"scheme"`
	TraceAddr     string `yaml:"traceAddr"`
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
