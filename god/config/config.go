package config

import (
	"encoding/json"
	"io/ioutil"
)

var gConfig Config

type MongoDBConfig struct {
	Url    string `json:"url"`
	DBName string `json:"dbname"`
}

type GtundConfig struct {
	Listener string `json:"gtund_listener"`
	Token    string `json:"token"` // 内部系统鉴权token
}

type GtunConfig struct {
	Listener string   `json:"gtun_listener"`
	Tokens   []string `json:"tokens"` // 用户授权码
}

type Config struct {
	Listener    string         `json:"listener"`
	GtundConfig *GtundConfig   `json:"gtund_config"`
	GtunConfig  *GtunConfig    `json:"gtun_config"`
	DBConfig    *MongoDBConfig `json:"database"`
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
	err := json.Unmarshal(content, &c)
	if err != nil {
		return nil, err
	}

	gConfig = c
	return &c, nil
}

func GetConfig() *Config {
	return &gConfig
}

func (c *Config) String() string {
	bytes, _ := json.MarshalIndent(c, "", "\t")
	return string(bytes)
}
