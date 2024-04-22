package fetcher

import (
	"encoding/json"
	"fmt"
)

var registerFetcher = map[string]Fetcher{}

type Fetcher interface {
	Name() string
	Setup(cfg json.RawMessage) error
	Fetch() (*FetchResult, error)
}

type FetchResult struct {
	Region  string   `json:"region"`
	GeoSite []string `json:"geo_site"`
	GeoIP   []string `json:"geo_ip"`
	Nodes   []Node   `json:"nodes"`
}

type Node struct {
	ServerAddr string `json:"server_addr"`
	TraceAddr  string `json:"trace_addr"`
	AuthKey    string `json:"auth_key"`
	Scheme     string `json:"scheme"`
	Rate       int    `json:"rate"`
}

func RegisterFetcher(fetcher Fetcher) {
	registerFetcher[fetcher.Name()] = fetcher
}

func GetFetcher(name string) Fetcher {
	return registerFetcher[name]
}

func Setup(name string, cfg json.RawMessage) error {
	f := registerFetcher[name]
	if f == nil {
		return fmt.Errorf("fetcher %s not register", name)
	}
	return f.Setup(cfg)
}
