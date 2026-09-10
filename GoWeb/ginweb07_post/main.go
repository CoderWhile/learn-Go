package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.LoadHTMLFiles("./login.html", "./index.html")

	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	}) //现在这个处理不了登录的post请求

	//login post
	r.POST("/login", func(c *gin.Context) {
		//获取form表单提交的数据
		//username := c.PostForm("username")
		//password := c.PostForm("password")//取到返回值,取不到返回空字符串

		//不填是返回空字符串
		//username := c.DefaultPostForm("username", "admin")
		//password := c.DefaultPostForm("XXX", "***")

		username, ok := c.GetPostForm("username")
		if !ok {
			username = "sb"
		}
		password, _ := c.GetPostForm("password")

		c.HTML(http.StatusOK, "index.html", gin.H{"Name": username, "Password": password})

	})
	r.Run(":8080")
}
