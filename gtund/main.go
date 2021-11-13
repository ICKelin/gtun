package gtund

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport"
	"github.com/ICKelin/gtun/transport/kcp"
	"github.com/ICKelin/gtun/transport/mux"
)

var version = ""

func init() {
	go http.ListenAndServe(":6060", nil)
}

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

	logs.Debug("config: %s", conf.String())

	var listener transport.Listener
	switch conf.ServerConfig.Scheme {
	case "kcp":
		listener = kcp.NewListener(conf.ServerConfig.Listen, []byte(conf.ListenerConfig))
		err := listener.Listen()
		if err != nil {
			logs.Error("new kcp server fail; %v", err)
			return
		}
		defer listener.Close()

	default:
		listener = mux.NewListener(conf.ServerConfig.Listen)
		err := listener.Listen()
		if err != nil {
			logs.Error("new mux server fail: %v", err)
			return
		}
		defer listener.Close()
	}

	server := NewServer(listener)
	server.Run()
}
