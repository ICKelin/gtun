package gtun

import (
	"flag"
	"fmt"

	"github.com/ICKelin/gtun/pkg/logs"
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

	client := NewClient(conf.ClientConfig)
	client.Run()
}
