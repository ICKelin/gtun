package gtun

import (
	"flag"
	"fmt"
)

func Main() {
	flgConf := flag.String("c", "", "config file")
	flag.Parse()

	conf, err := ParseConfig(*flgConf)
	if err != nil {
		fmt.Printf("load config fail: %v\n", err)
		return
	}

	registry := NewRegistry(conf.RegistryConfig)

	client := NewClient(conf.ClientConfig, registry)
	client.Run()
}
