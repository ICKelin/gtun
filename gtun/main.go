package gtun

import (
	"flag"
	"fmt"

	"github.com/ICKelin/gtun/pkg/logs"
)

func Main() {
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Printf("load config fail: %v\n", err)
		return
	}
	logs.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)

	tcpfw := NewTCPForward(conf.TCPForward)
	lis, err := tcpfw.Listen()
	if err != nil {
		logs.Error("listen tproxy tcp fail: %v", err)
		return
	}

	go tcpfw.Serve(lis)

	udpfw := NewUDPForward(conf.UDPForward)
	udpConn, err := udpfw.Listen()
	if err != nil {
		logs.Error("listen tproxy udp fail: %v", err)
		return
	}

	go udpfw.Serve(udpConn)

	client := NewClient(conf.ClientConfig)
	client.Run()
}
