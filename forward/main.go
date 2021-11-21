package forward

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport/transport_api"
)

func Main() {
	flgConf := flag.String("c", "", "config file path")
	flag.Parse()

	cfg, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	logs.Init("forward.log", "debug", 10)
	logs.Debug("forward config: %v", cfg)

	// initial local listener
	lisCfg := cfg.ListenerConfig
	listener, err := transport_api.NewListen(lisCfg.Scheme, lisCfg.ListenAddr, lisCfg.RawConfig)
	if err != nil {
		logs.Error("new listener fail: %v", err)
		return
	}
	defer listener.Close()

	// initial next hop dialer
	dialerCfg := cfg.NexthopConfig
	routeTable := NewRouteTable()
	err = routeTable.Add(dialerCfg.Scheme, dialerCfg.NexthopAddr, dialerCfg.RawConfig)
	if err != nil {
		logs.Error("add route table fail: %v", err)
		return
	}

	f := NewForward(listener, routeTable)

	if err := f.Serve(); err != nil {
		logs.Error("forward exist: %v", err)
	}
}
