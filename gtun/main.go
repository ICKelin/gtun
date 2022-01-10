package gtun

import (
	"flag"
	"fmt"
	"github.com/ICKelin/optw/transport/transport_api"
	"net/http"
	_ "net/http/pprof"

	"github.com/ICKelin/gtun/internal/logs"
)

func init() {
	go http.ListenAndServe(":6060", nil)
}

func Main() {
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Printf("load config fail: %v\n", err)
		return
	}
	logs.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)

	for region, cfg := range conf.Forwards {
		dialer, err := transport_api.NewDialer(cfg.Transport.Scheme, cfg.ServerAddr, cfg.Transport.ConfigContent)
		if err != nil {
			logs.Error("new dialer fail: %v", err)
			continue
		}

		tcpfw := NewTCPForward(region, cfg.TCPForward)
		lis, err := tcpfw.Listen()
		if err != nil {
			logs.Error("listen tproxy tcp fail: %v", err)
			return
		}

		go tcpfw.Serve(lis)

		udpfw := NewUDPForward(region, cfg.UDPForward)
		udpConn, err := udpfw.Listen()
		if err != nil {
			logs.Error("listen tproxy udp fail: %v", err)
			return
		}

		go udpfw.Serve(udpConn)

		client := NewClient(dialer)
		go client.Run(region)

	}

	select {}
}
