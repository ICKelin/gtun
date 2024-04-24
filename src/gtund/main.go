package main

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/src/internal/logs"
	"github.com/ICKelin/optw/transport/transport_api"
	"net/http"
	_ "net/http/pprof"
	"time"
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

		if conf.EnableAuth {
			listener.SetAuthFunc(func(token string) bool {
				ok := false
				for _, auth := range conf.Auths {
					if auth.AccessToken == token {
						if auth.ExpiredAt == 0 {
							ok = true
						} else if time.Now().Unix() < auth.ExpiredAt {
							ok = true
						}
						break
					}
				}
				return ok
			})
		}

		s := NewServer(listener)
		go s.Run()
	}

	select {}
}
