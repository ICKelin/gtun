package main

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/src/gtun/config"
	"github.com/ICKelin/gtun/src/gtun/proxy"
	"github.com/ICKelin/gtun/src/gtun/route"
	"github.com/ICKelin/gtun/src/internal/logs"
)

func main() {
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	conf, err := config.Parse(*flgConf)
	if err != nil {
		fmt.Printf("load config fail: %v\n", err)
		return
	}
	logs.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)

	routeConfig, err := config.ParseRoute(conf.RouteFile)
	if err != nil {
		fmt.Printf("parse node config fail: %v", err)
		return
	}

	proxyConfig, err := config.ParseProxy(conf.ProxyFile)
	if err != nil {
		fmt.Printf("parse proxy config fail: %v", err)
		return
	}

	// run route
	err = route.Setup(routeConfig)
	if err != nil {
		fmt.Printf("route setup fail: %v", err)
		return
	}

	// run proxy
	err = proxy.Serve(proxyConfig)
	if err != nil {
		fmt.Printf("proxy setup fail: %v", err)
		return
	}
	// TODO: watch for config file changes
	select {}
}
