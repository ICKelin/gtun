package gtun

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/gtun/proxy"
	"github.com/ICKelin/gtun/gtun/route"
	"github.com/ICKelin/gtun/internal/logs"
)

var (
	needSysInit = false
	httpListen  = ":9095"
	confPath    = ""
)

func Main() {
	flag.StringVar(&confPath, "c", "", "config file")
	flag.Parse()

	conf, err := ParseConfig(confPath)
	if err != nil {
		fmt.Printf("load config fail: %v, waiting for sys init\n", err)
		needSysInit = true
		panic(NewHTTPServer(httpListen).ListenAndServe())
		return
	}
	logs.Init(conf.Log.Path, conf.Log.Level, conf.Log.Days)

	// run proxy
	for region, cfg := range conf.Settings {
		// init plugins
		err = proxy.GetManager().Setup(region, cfg.RegionProxyFile, cfg.IPProxyFile, cfg.Proxy)
		if err != nil {
			fmt.Printf("set proxy fail: %v\n", err)
			return
		}
	}

	// run route and race
	raceManager := route.GetTraceManager()
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

		regionRace := route.NewTrace(region, raceTargets)
		raceManager.AddRegionTrace(region, regionRace)
	}
	raceManager.RunRace()
	panic(NewHTTPServer(httpListen).ListenAndServe())
}
