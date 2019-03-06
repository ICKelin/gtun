package god

import (
	"encoding/json"

	"github.com/ICKelin/gtun/god/config"
	"github.com/ICKelin/gtun/god/registry"
	"github.com/ICKelin/gtun/logs"
	"github.com/gin-gonic/gin"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		logs.Error(err)
		return
	}

	config, err := config.ParseConfig(opts.confPath)
	if err != nil {
		logs.Error("parse config error: %v", err)
		return
	}

	go func() {
		d := registry.NewGtund(config.GtundConfig)
		logs.Error(d.Run())
	}()

	go func() {
		c := registry.NewGtun(config.GtunConfig)
		logs.Error(c.Run())
	}()

	engine := gin.Default()
	engine.GET("/status", status)
	err = engine.Run(config.Listener)
	if err != nil {
		logs.Error(err)
	}
}

func status(ctx *gin.Context) {
	results := registry.GetDB().GtundList()
	bytes, _ := json.Marshal(results)
	ctx.Writer.Write(bytes)
}
