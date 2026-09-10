package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/index", func(c *gin.Context) {
		//c.JSON(http.StatusOK, gin.H{
		//	"status": "ok",
		//})
		//跳转到baidu.com
		c.Redirect(http.StatusMovedPermanently, "http://www.baidu.com/")
		//301永久重定向,浏览器会永久记录
	})

	//路由重定向:
	r.GET("/a", func(c *gin.Context) {
		//跳转到b对应的路由函数
		c.Request.URL.Path = "/b" //把请求的uri修改
		r.HandleContext(c)        //继续后续的处理
	})
	//都执行了一遍
	r.GET("/b", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "b",
		})
	})
	r.Run(":8080")

}
