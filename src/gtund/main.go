package main

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/src/internal/logs"
	"net/http"
	_ "net/http/pprof"

	"github.com/ICKelin/optw/transport/transport_api"
)

var version = ""

func init() {
	go http.ListenAndServe(":6060", nil)
}

func main() {
	flgVersion := flag.Bool("v", false, "print version")
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	if *flgVersion {
		fmt.Println(version)
		return
	}

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Printf("parse config file fail: %s %v\n", *flgConf, err)
		return
	}
	logs.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)

	logs.Debug("config: %s", conf.String())

	if conf.Trace != "" {
		go NewTraceServer(conf.Trace).ListenAndServe()
	}

	for _, cfg := range conf.ServerConfig {
		listener, err := transport_api.NewListen(cfg.Scheme, cfg.Listen, cfg.ListenerConfig)
		if err != nil {
			logs.Error("new listener fail: %v", err)
			return
		}
		defer listener.Close()

		s := NewServer(listener)
		go s.Run()
	}

	select {}
}
