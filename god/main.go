package god

import (
	"github.com/ICKelin/glog"
	"github.com/ICKelin/gtun/common"
	"github.com/gin-gonic/gin"
)

func Main() {
	opts, err := ParseArgs()
	if err != nil {
		glog.FATAL(err)
	}

	config, err := ParseConfig(opts.confPath)
	if err != nil {
		glog.FATAL(err)
	}

	go func() {
		d := NewGtund(config.GtundConfig)
		glog.FATAL(d.Run())
	}()

	go func() {
		c := NewGtun(config.GtunConfig)
		glog.FATAL(c.Run())
	}()

	engine := gin.Default()

	engine.POST("/gtund/report", func(ctx *gin.Context) {
		var reg = common.S2GRegister{}
		if err := ctx.BindJSON(&reg); err != nil {
			return
		}
	})

	engine.POST("/gtun/report", func(ctx *gin.Context) {

	})

	err = engine.Run(config.Listener)
	if err != nil {
		glog.ERROR(err)
	}
}
