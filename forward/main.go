package forward

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/internal/logs"
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

	for _, routeCfg := range cfg.RouteConfig {
		// initial local listener
		lisCfg := routeCfg.ListenerConfig

		// initial next hop dialer
		dialerCfg := routeCfg.NexthopConfig
		routeTable := NewRouteTable()
		err = routeTable.Add(dialerCfg.Scheme, dialerCfg.NexthopAddr, dialerCfg.RawConfig)
		if err != nil {
			logs.Error("add route table fail: %v", err)
			return
		}

		f := NewForward(lisCfg.ListenAddr, routeTable)

		if err := f.ListenAndServe(); err != nil {
			logs.Error("forward exist: %v", err)
		}
	}

	select {}
}
