package registry

import (
	"flag"
	"net/http"

	"github.com/ICKelin/gtun/logs"
)

func Main() {
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		logs.Error("parse config error: %v", err)
		return
	}

	m := NewModels()

	d := NewGtund(conf.GtundConfig, m)
	go func() {
		logs.Error("run server for gtund fail: %v", d.Run())
	}()

	g := NewGtun(conf.GtunConfig, m)

	logs.Info("api listen %s", g.listener)

	http.HandleFunc("/gtun/access", g.onGtunAccess)
	http.HandleFunc("/gtund/list", d.GetGtundList)
	http.ListenAndServe(g.listener, nil)
}
