package main

import (
	"flag"
	"github.com/ICKelin/gtun/src/agent/fetcher"
)

var (
	confPath = ""
)

func main() {
	flag.StringVar(&confPath, "c", "", "config file path")
	flag.Parse()

	conf, err := ParseConfig(confPath)
	if err != nil {
		panic(err)
	}

	for name, cfg := range conf.Fetcher {
		err := fetcher.Setup(name, cfg)
		if err != nil {
			panic(err)
		}
	}

	tagManager, _ := NewTagManager(conf.GeoConfig.GeoIPFile, conf.GeoConfig.GeoDomainFile)
	daemon := NewDaemon(conf.GtunConfig, tagManager)
	daemon.WatchGtun()
}
