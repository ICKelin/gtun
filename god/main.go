package god

import (
	"encoding/json"

	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/god/config"
	"github.com/ICKelin/gtun/god/registry"
	"github.com/gin-gonic/gin"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		glog.FATAL(err)
	}

	config, err := config.ParseConfig(opts.confPath)
	if err != nil {
		glog.FATAL(err)
	}

	go func() {
		d := registry.NewGtund(config.GtundConfig)
		glog.FATAL(d.Run())
	}()

	go func() {
		c := registry.NewGtun(config.GtunConfig)
		glog.FATAL(c.Run())
	}()

	engine := gin.Default()
	engine.GET("/status", status)
	err = engine.Run(config.Listener)
	if err != nil {
		glog.ERROR(err)
	}
}

func status(ctx *gin.Context) {
	results := registry.GetDB().GtundList()
	bytes, _ := json.Marshal(results)
	ctx.Writer.Write(bytes)
}
