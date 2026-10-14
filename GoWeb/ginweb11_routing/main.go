package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//路由与路由组

func main() {
	r := gin.Default()
	//访问/index的GET请求会走这一条处理逻辑
	//路由
	//get-获取信息
	r.GET("/index", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"method": "GET",
		})
	})

	//上传文件,填表单-创建操作(注册)
	r.POST("/index", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"method": "POST",
		})
	})

	//更新(修改个人信息部分数据)
	r.PUT("/index", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"method": "PUT",
		})
	})

	r.DELETE("/index", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"method": "DELETE",
		})
	})
	//请求方法大集合大杂烩
	r.Any("/user", func(c *gin.Context) {
		switch c.Request.Method {
		case "GET":
			c.JSON(http.StatusOK, gin.H{
				"method": "GET",
			})
		case "POST":
			c.JSON(http.StatusOK, gin.H{
				"method": "POST",
			})

		}

		c.JSON(http.StatusOK, gin.H{
			"method": "Any",
		})
	})

	//NoRoute
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "lele",
		})
	})

	//视频的首页和详情页
	//路由组:
	videoGroup := r.Group("/video")
	{
		videoGroup.GET("/index", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"msg": "/video/index",
			})
		})

		videoGroup.POST("/index", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"msg": "/video/index",
			})
		})
	}
	//r.GET("video/index", func(c *gin.Context) {
	//	c.JSON(http.StatusOK, gin.H{
	//		"msg": "/video/index"})
	//})

	r.Run(":8080")
}
