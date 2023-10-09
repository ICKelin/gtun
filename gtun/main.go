package gtun

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/gtun/proxy"
	"github.com/ICKelin/gtun/gtun/route"
	"github.com/ICKelin/gtun/internal/logs"
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
			hopConn, err := route.CreateConnection(region, r.Scheme, r.Addr, r.AuthKey)
			if err != nil {
				fmt.Printf("connect to %s://%s fail: %v\n", r.Scheme, r.Addr, err)
				return
			}
			go hopConn.ConnectNextHop()
		}

		regionRace := route.NewRace(region, raceTargets)
		raceManager.AddRegionRace(region, regionRace)
	}
	raceManager.RunRace()

	select {}
}
