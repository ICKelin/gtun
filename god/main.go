package god

import (
	"github.com/ICKelin/glog"
	"github.com/gin-gonic/gin"
)

func Main() {
	engine := gin.Default()
	err := engine.Run(":9622")
	if err != nil {
		glog.ERROR(err)
	}
}
