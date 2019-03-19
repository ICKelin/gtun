package registry

import (
	"github.com/ICKelin/gtun/logs"
	"github.com/ICKelin/gtun/registry/config"
	"github.com/ICKelin/gtun/registry/server"
	"github.com/gin-gonic/gin"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		logs.Error("parse args fail: %v", err)
		return
	}

	config, err := config.ParseConfig(opts.confPath)
	if err != nil {
		logs.Error("parse config error: %v", err)
		return
	}

	go func() {
		d := server.NewGtund(config.GtundConfig)
		logs.Error("run server for gtund fail: %v", d.Run())
	}()

	go func() {
		c := server.NewGtun(config.GtunConfig)
		logs.Error("run api for gtun fail: %v", c.Run())
	}()

	engine := gin.Default()
	engine.GET("/status", status)
	err = engine.Run(config.Listener)
	if err != nil {
		logs.Error("run api for gtun fail: %v", err)
	}
}

func status(ctx *gin.Context) {
	// results := models.GetDB().
	// bytes, _ := json.Marshal(results)
	// ctx.Writer.Write(bytes)
}
