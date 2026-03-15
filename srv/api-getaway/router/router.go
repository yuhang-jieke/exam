package router

import (
	"github.com/gin-gonic/gin"
	"github.com/yuhang-jieke/exam/srv/api-getaway/handler/server"
)

func Router() *gin.Engine {
	r := gin.Default()
	r.POST("/goods/add", server.GoodsAdd)
	r.POST("/pay", server.AliPay)
	return r
}
