package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	r.GET("/web", func(c *gin.Context) {
		//获取浏览器那边发请求携带的query string参数
		//name := c.Query("query") //通过query获取请求中携带的querystring参数
		//娶不到用指定的默认值
		//name := c.DefaultQuery("query", "sombody")
		name, flag := c.GetQuery("query")

		if !flag {
			//取不到
			name = "somebody"
		}
		c.JSON(200, gin.H{
			"name": name,
		})
	})

	r.Run(":8080")
}
