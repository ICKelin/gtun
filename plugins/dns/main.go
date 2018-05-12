package dns

import (
	_ "net/http/pprof"

	"github.com/ICKelin/glog"
)

func Run(confpath string) {
	conf, err := LoadConfig(confpath)
	if err != nil {
		glog.ERROR(err)
		return
	}

	selfDefine := NewSelfDefine(conf.RulesDir, conf.RulesDir+"/config")
	go selfDefine.Run()

	resolver := NewResolver()

	worker := NewWorker(conf.Workers, conf.BufferSize, selfDefine, resolver)
	go worker.Run()

	dns := NewDNS(":53", worker)
	go dns.Run()
}
