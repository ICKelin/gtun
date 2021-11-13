package forward

import (
	"flag"
	"fmt"
	"github.com/ICKelin/gtun/internal/logs"
	"github.com/ICKelin/gtun/transport"
	"github.com/ICKelin/gtun/transport/kcp"
	"github.com/ICKelin/gtun/transport/mux"
)

func Main() {
	flgConf := flag.String("c", "", "config file path")
	flag.Parse()

	cfg, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Println(err)
		return
	}

	// initial local listener
	var listener transport.Listener
	lisCfg := cfg.ListenerConfig
	switch lisCfg.Scheme {
	case "kcp":
		listener = kcp.NewListener(lisCfg.ListenAddr, []byte(lisCfg.RawConfig))
		err := listener.Listen()
		if err != nil {
			logs.Error("new kcp server fail; %v", err)
			return
		}
		defer listener.Close()

	default:
		listener := mux.NewListener(lisCfg.ListenAddr)
		err := listener.Listen()
		if err != nil {
			logs.Error("new mux server fail: %v", err)
			return
		}
		defer listener.Close()
	}

	// initial nexthop dialer
	var dialer transport.Dialer
	dialerCfg := cfg.NexthopConfig
	switch dialerCfg.Scheme {
	case "kcp":
		dialer = kcp.NewDialer(dialerCfg.NexthopAddr, []byte(dialerCfg.RawConfig))
	default:
		dialer = mux.NewDialer(dialerCfg.NexthopAddr)
	}

	f := NewForward(listener, dialer)

	if err := f.Serve(); err != nil {
		logs.Error("forward exist: %v", err)
	}
}
