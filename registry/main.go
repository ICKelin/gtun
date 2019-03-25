package registry

import (
	"flag"

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

	go func() {
		d := NewGtund(conf.GtundConfig)
		logs.Error("run server for gtund fail: %v", d.Run())
	}()

	c := NewGtun(conf.GtunConfig)
	logs.Error("run api for gtun fail: %v", c.Run())
}
