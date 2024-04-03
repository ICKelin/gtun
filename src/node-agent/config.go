package main

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	GtunTemplateFile      string `yaml:"gtun_template_file"`
	GtunService           string `yaml:"gtun_service"`
	GtunConfigFilePath    string `yaml:"gtun_config_file_path"`
	GtunDynamicConfigFile string `yaml:"gtun_dynamic_config_file"`
	Log                   Log    `yaml:"log"`
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
