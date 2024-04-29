package main

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/src/gtun/config"
	"github.com/ICKelin/gtun/src/gtun/proxy"
	"github.com/ICKelin/gtun/src/gtun/route"
	"github.com/ICKelin/gtun/src/internal/logs"
)

var logo = `
====================================
 ██████  ████████ ██    ██ ███    ██ 
██          ██    ██    ██ ████   ██ 
██   ███    ██    ██    ██ ██ ██  ██ 
██    ██    ██    ██    ██ ██  ██ ██ 
 ██████     ██     ██████  ██   ████ 
====================================
https://github.com/ICKelin/gtun
`

func main() {
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	conf, err := config.Parse(*flgConf)
	if err != nil {
		fmt.Printf("load config fail: %v\n", err)
		return
	}
	fmt.Println(logo)
	logs.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)
	logs.Info("%s", logo)
	for region, cfg := range conf.Accelerator {
		err := route.Setup(region, cfg.Routes)
		if err != nil {
			panic(err)
		}

		err = proxy.Serve(region, cfg.Proxy)
		if err != nil {
			panic(err)
		}
	}
	route.Run()

	// TODO: watch for config file changes
	select {}
}
