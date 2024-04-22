package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ICKelin/gtun/src/agent/fetcher"
	"github.com/beyond-net/golib/logs"
	"os"
	"os/exec"
	"text/template"
	"time"
	"v2ray.com/core/app/router"
)

type routeItem struct {
	ServerAddr string `json:"server_addr"`
	TraceAddr  string `json:"trace_addr"`
	AuthKey    string `json:"auth_key"`
	Scheme     string `json:"scheme"`
	Rate       int    `json:"rate"`
}

type proxyObj struct {
	Region     string
	ListenPort int
}

var proxyTemplate = `
"{{.Region}}":
  tproxy_tcp: |
    {
      "read_timeout": 30,
      "write_timeout": 30,
      "listen_addr": ":{{.ListenPort}}"
    }
  tproxy_udp: |
    {
      "read_timeout": 30,
      "write_timeout": 30,
      "session_timeout": 30,
      "listen_addr": ":{{.ListenPort}}"
    }

`

type Daemon struct {
	siteList   []router.GeoSite
	ipList     []router.GeoIP
	gtunConfig *GtunConfig
	tagManager *TagManager
}

func NewDaemon(cfg *GtunConfig, tagManager *TagManager) *Daemon {
	return &Daemon{gtunConfig: cfg, tagManager: tagManager}
}

func (daemon *Daemon) WatchGtun() {
	f := fetcher.GetFetcher(daemon.gtunConfig.FetcherName)
	if f == nil {
		logs.Error("invalid fetcher %s", daemon.gtunConfig.FetcherName)
		return
	}

	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()
	for range tick.C {
		rs, err := f.Fetch()
		if err != nil {
			logs.Warn("fetcher[%s] fail: %v", f.Name(), err)
			continue
		}

		logs.Debug("fetcher[%s] result: %v", f.Name(), rs)
		err = daemon.reloadGtun(rs)
		if err != nil {
			logs.Warn("reload gtun fail: %v", err)
			continue
		}
	}
}

func (daemon *Daemon) reloadGtun(newCfg *fetcher.FetchResult) error {
	routes, proxyContent, err := daemon.render(newCfg)
	if err != nil {
		return err
	}

	// generate ip/domains
	ips, domains, err := daemon.geoInfo(newCfg.GeoIP, newCfg.GeoSite)
	if err != nil {
		return err
	}

	logs.Debug(ips, domains)

	// write routes.json
	routesBytes, err := json.Marshal(routes)
	if err != nil {
		return err
	}
	routeFp, err := os.Open(daemon.gtunConfig.RouteFile)
	if err != nil {
		return err
	}
	defer routeFp.Close()
	_, err = routeFp.Write(routesBytes)
	if err != nil {
		return err
	}

	// write proxy.yaml
	fp, err := os.Open(daemon.gtunConfig.ProxyFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	_, err = fp.Write([]byte(proxyContent))
	if err != nil {
		return err
	}

	return daemon.restartGtun()
}

func (daemon *Daemon) render(newCfg *fetcher.FetchResult) ([]*routeItem, string, error) {
	routes := make([]*routeItem, 0)
	proxyContent := ""
	basePort := 8154
	for _, node := range routes {
		// render route file
		routes = append(routes, &routeItem{
			Scheme:     node.Scheme,
			TraceAddr:  node.TraceAddr,
			AuthKey:    node.AuthKey,
			ServerAddr: node.ServerAddr,
			Rate:       node.Rate,
		})

		// render proxy file
		tpl := template.New("proxy")
		tpl, err := tpl.Parse(proxyTemplate)
		if err != nil {
			return nil, "", err
		}

		br := &bytes.Buffer{}
		err = tpl.Execute(br, &proxyObj{
			Region:     newCfg.Region,
			ListenPort: basePort,
		})
		if err != nil {
			return nil, "", err
		}
		proxyContent += string(br.Bytes())
		basePort += 1
	}

	return routes, proxyContent, nil
}

func (daemon *Daemon) geoInfo(geoIP []string, geoDomain []string) ([]string, []string, error) {
	ips := make([]string, 0)
	domains := make([]string, 0)

	for _, region := range geoIP {
		ips = append(ips, daemon.tagManager.GetTagIPList(region)...)
	}

	for _, site := range geoDomain {
		domains = append(domains, daemon.tagManager.GetTagIPList(site)...)
	}
	return ips, domains, nil
}

func (daemon *Daemon) restartGtun() error {
	command := daemon.gtunConfig.RestartCmd
	if len(command) <= 0 {
		return fmt.Errorf("invalid cmd")
	}
	_, err := exec.Command(command[0], command[1:]...).CombinedOutput()
	if err != nil {
		logs.Warn("%v", err)
	}
	return err
}
