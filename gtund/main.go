package gtund

import (
	"flag"
	"fmt"

	"github.com/ICKelin/gtun/pkg/logs"
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

	dhcp, err := NewDHCP(conf.DHCPConfig)
	if err != nil {
		logs.Error("init dhcp fail: %v", err)
		return
	}

	iface, err := NewInterface(conf.IsTap, conf.DHCPConfig.Gateway, conf.DHCPConfig.CIDR)
	if err != nil {
		logs.Error("init interface fail: %v", err)
		return
	}

	server, err := NewServer(conf.ServerConfig, dhcp, iface)
	if err != nil {
		logs.Error("new server: %v", err)
		return
	}

	server.Run()
}
