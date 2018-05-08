package main

import (
	"github.com/ICKelin/glog"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	err := engine.Run(":9622")
	if err != nil {
		glog.ERROR(err)
	}
}
