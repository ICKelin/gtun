package gtund

import (
	"flag"
	"fmt"

	"github.com/ICKelin/gtun/internal/logs"
)

var version = ""

func Main() {
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

	server, err := NewServer(conf.ServerConfig)
	if err != nil {
		logs.Error("new server: %v", err)
		return
	}

	server.Run()
}
