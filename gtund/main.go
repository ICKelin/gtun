package gtund

import (
	"fmt"

	"github.com/ICKelin/glog"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Printf("parse args fail: %v", err)
		return
	}

	if opts.debug {
		glog.Init("gtund", glog.PRIORITY_DEBUG, "./", glog.OPT_DATE, 1024*10)
	} else {
		glog.Init("gtund", glog.PRIORITY_WARN, "./", glog.OPT_DATE, 1024*10)
	}

	var conf *Config
	if opts.confpath != "" {
		conf, err = ParseConfig(opts.confpath)
		if err != nil {
			glog.FATAL("parse config file fail:", opts.confpath, err)
		}
	}

	if conf != nil && conf.GodCfg != nil {
		go func() {
			g := NewGod(conf.GodCfg)
			glog.FATAL(g.Run())
		}()
	}

	serverCfg := &ServerConfig{
		listenAddr:  opts.listenAddr,
		authKey:     opts.authKey,
		gateway:     opts.gateway,
		routeUrl:    opts.routeUrl,
		nameservers: opts.nameserver,
		reverseFile: opts.reverseFile,
	}

	server, err := NewServer(serverCfg)
	if err != nil {
		glog.ERROR(err)
		return
	}

	server.Run()
	server.Stop()
}
