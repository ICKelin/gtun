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

	for _, cfg := range conf.Forwards {
		tcpfw := NewTCPForward(cfg.Region, cfg.TCPForward)
		lis, err := tcpfw.Listen()
		if err != nil {
			logs.Error("listen tproxy tcp fail: %v", err)
			return
		}

		go tcpfw.Serve(lis)

		udpfw := NewUDPForward(cfg.Region, cfg.UDPForward)
		udpConn, err := udpfw.Listen()
		if err != nil {
			logs.Error("listen tproxy udp fail: %v", err)
			return
		}

		go udpfw.Serve(udpConn)

		for _, hopCfg := range cfg.Transport {
			dialer, err := transport_api.NewDialer(hopCfg.Scheme, hopCfg.Server, hopCfg.ConfigContent)
			if err != nil {
				logs.Error("new dialer fail: %v", err)
				continue
			}

			client := NewClient(dialer)
			go client.Run(cfg.Region)
		}
	}

	select {}
}
