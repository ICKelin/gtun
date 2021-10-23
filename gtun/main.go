package gtun

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

		var dialer transport.Dialer
		switch cfg.Transport.Scheme {
		case "kcp":
			dialer = kcp.NewDialer([]byte(cfg.Transport.ConfigContent))
		default:
			dialer = &mux.Dialer{}
		}

		client := NewClient(dialer)
		go client.Run(region, cfg.ServerAddr)
	}

	select {}
}
