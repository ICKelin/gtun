package gtund

import (
	"flag"
	"fmt"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport"
	"github.com/ICKelin/gtun/transport/kcp"
	"github.com/ICKelin/gtun/transport/mux"
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

	var listener transport.Listener
	switch conf.ServerConfig.Scheme {
	case "kcp":
		listener, err = kcp.Listen(conf.ServerConfig.Listen)
		if err != nil {
			fmt.Printf("new kcp server fail: %v", err)
		}
		defer listener.Close()

	default:
		listener, err = mux.Listen(conf.ServerConfig.Listen)
		if err != nil {
			fmt.Printf("new mux server fail: %v", err)
		}
		defer listener.Close()
	}

	server := NewServer(listener)
	server.Run()
}
