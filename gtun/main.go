package gtun

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/gtun/proxy"
	"github.com/ICKelin/gtun/gtun/route"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/optw/transport/transport_api"
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

	// run proxy
	for _, cfg := range conf.Settings {
		// init plugins
		err = proxy.Setup(cfg.Proxy)
		if err != nil {
			fmt.Printf("set proxy fail: %v", err)
			return
		}
	}

	// run route and race
	raceManager := route.GetRaceManager()
	for region, cfg := range conf.Settings {
		raceTargets := make([]string, 0)
		for _, r := range cfg.Route {
			raceTargets = append(raceTargets, r.TraceAddr)
			dialer, err := transport_api.NewDialer(r.Scheme, r.Addr, "")
			if err != nil {
				fmt.Printf("new dialer fail: %v", err)
				return
			}

			raceTargets = append(raceTargets, r.TraceAddr)
			go route.NewClient(region, dialer).ConnectNextHop()
		}

		regionRace := route.NewRace(raceTargets)
		raceManager.AddRegionRace(region, regionRace)
	}
	raceManager.RunRace()

	select {}
}
