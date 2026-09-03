package main

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

//func main() {
//	r := gin.Default()
//	r.LoadHTMLGlob("templates/**/*")
//	//r.LoadHTMLFiles("templates/posts/index.html", "templates/users/index.html")
//	r.GET("/posts/index", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "posts/index.html", gin.H{
//			"title": "posts/index",
//		})
//	})
//
//	r.GET("users/index", func(c *gin.Context) {
//		c.HTML(http.StatusOK, "users/index.html", gin.H{
//			"title": "users/index",
//		})
//	})
//
//	r.Run(":8080")
//}

//gin框架

//静态文件:
//html页面上用到的样式文件,css.js文件,图片

func main() {
	r := gin.Default()
	//加载静态文件
	r.Static("/xxx", "./statics")
	//自定义一个函数加载到模板里
	r.SetFuncMap(template.FuncMap{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	})
	//一个默认的路由
	//r.LoadHTMLGlob("templates/**/*")
	r.LoadHTMLFiles("templates/posts/index.html", "templates/users/index.html") //模板解析
	r.GET("/posts/index", func(c *gin.Context) {
		//gin.H:本质就是一个map
		c.HTML(http.StatusOK, "posts/index.html", gin.H{ //模板渲染
			"title": "posts/index", //返回一个模板文件并渲染数据
		})
	})
	r.GET("/users/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "users/index.html", gin.H{
			"title": "<a href='https://liwenzhou.com'>李文周</a>",
		})

	})
	r.Run(":8080") //启动sever
}
