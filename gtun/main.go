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

	raceTargets := make(map[string][]string)
	for region, hops := range conf.Route {
		for _, hop := range hops {
			raceTargets[region] = append(raceTargets[region], hop.TraceAddr)
			dialer, err := transport_api.NewDialer(hop.Scheme, hop.Addr, "")
			if err != nil {
				fmt.Printf("new dialer fail: %v", err)
				return
			}

			go route.NewClient(region, dialer).ConnectNextHop()
		}
	}

	// init race
	raceManager := route.GetRaceManager()
	for region, targets := range raceTargets {
		race := route.NewRace(targets)
		raceManager.AddRegionRace(region, race)
	}

	// init plugins
	err = proxy.Setup(conf.Proxy)
	if err != nil {
		fmt.Printf("set proxy fail: %v", err)
		return
	}

	select {}
}
