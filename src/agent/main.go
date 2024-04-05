package main

import (
	"flag"
	"fmt"
	"github.com/beyond-net/golib/logs"
	"github.com/radovskyb/watcher"
	"time"
)

var (
	confPath = ""
)

func main() {
	flag.StringVar(&confPath, "c", "", "config file path")
	flag.Parse()

	cfg, err := ParseConfig(confPath)
	if err != nil {
		fmt.Printf("parse config fail: %v\n", err)
		return
	}

	logs.Init(cfg.Log.Path, cfg.Log.Level, 5)
	logs.Info("%+v", cfg)

	w := watcher.New()
	err = w.Add(cfg.GtunDynamicConfigFile)
	if err != nil {
		logs.Error("watch %s fail: %v", cfg.GtunDynamicConfigFile)
		return
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				logs.Info("file %s modify", event.FileInfo.Name())
				switch event.FileInfo.Name() {
				case cfg.GtunDynamicConfigFile:
					// TODO: reload gtun
				default:

				}
			case err := <-w.Error:
				logs.Warn("file watcher error occurs: %v", err)
			case <-w.Closed:
				return
			}
		}
	}()

	err = w.Start(time.Millisecond * 100)
	if err != nil {
		logs.Warn("file watcher start error: %v", err)
	}
	select {}
}
