package main

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/src/internal/logs"
	"github.com/ICKelin/optw/transport_api"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var version = ""

var logo = `
====================================
 ██████  ████████ ██    ██ ███    ██ 
██          ██    ██    ██ ████   ██ 
██   ███    ██    ██    ██ ██ ██  ██ 
██    ██    ██    ██    ██ ██  ██ ██ 
 ██████     ██     ██████  ██   ████ 
====================================
https://github.com/ICKelin/gtun`

func init() {
	go http.ListenAndServe(":6060", nil)
}

func main() {
	flgVersion := flag.Bool("v", false, "print version")
	flgTest := flag.Bool("t", false, "test config file")
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	fmt.Println(logo)

	if *flgVersion {
		fmt.Println(version)
		return
	}

	if *flgTest {
		_, err := ParseConfig(*flgConf)
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
		}
		return
	}

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Printf("parse config file fail: %s %v\n", *flgConf, err)
		return
	}
	logs.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)
	logs.Info("%s", logo)
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
				auths := GetConfig().Auths
				logs.Info("verify auth token %s", token)
				for _, auth := range auths {
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	for sig := range c {
		switch sig {
		case syscall.SIGHUP:
			logs.Info("receive hup signal")
			cfg, err := ParseConfig(*flgConf)
			if err != nil {
				logs.Warn("reload config fail: %v", err)
				continue
			}

			// rewrite conf pointer
			conf = cfg
			logs.Info("reload config success")

		default:
			logs.Info("un handle signal %v", sig.String())
		}
	}
}
