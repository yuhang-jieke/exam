package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuhang-jieke/exam/srv/api-getaway/basic/config"
	__ "github.com/yuhang-jieke/exam/srv/proto"
)

func GoodsAdd(c *gin.Context) {
	var form __.AddGoodsReq
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数不正确",
		})
		return
	}

	// Contact the server and print out its response.
	_, err := config.GoodsClient.AddGoods(c, &__.AddGoodsReq{
		Name:  form.Name,
		Price: form.Price,
		Stock: form.Stock,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "添加失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "添加成功",
	})
	return
}
func AliPay(c *gin.Context) {
	var form __.PayReq
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "参数不正确",
		})
		return
	}

	// Contact the server and print out its response.
	r, err := config.GoodsClient.AliPay(c, &__.PayReq{
		Id: form.Id,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"msg":  "添加失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"url":  r,
	})
	return
}
