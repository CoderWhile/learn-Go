package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// 中间件
// 属于HandlerFunc类型,也可以称为一个中间件
func indexHandler(c *gin.Context) {
	fmt.Println("index")
	name, ok := c.Get("name")
	if !ok {
		name = "匿名用户"
	}
	c.JSON(200, gin.H{
		"msg":  "index",
		"name": name,
	})
}

// 定义一个中间件m1:统计耗时
func m1(c *gin.Context) {
	fmt.Println("m1 in...")
	//计时
	start := time.Now()
	c.Next() //调用后续的处理函数
	//c.Abort() //阻止调用后续的处理函数

	cost := time.Since(start)
	fmt.Printf("cost:%v\n", cost)
	fmt.Println("m1 out.")
}

func m2(c *gin.Context) {
	fmt.Println("m2 in ..")
	c.Set("name", "qimi") //想让index函数拿到值

	//c.Abort()//阻止调用后续的处理函数
	//return
	fmt.Println("m2 out...")
}

//func authMiddleware(c *gin.Context) {
//	//是否登录的判断
//	//if 是登录用户
//	//c.Next()
//	//else
//	//c.Abort()
//
//}

func authMiddleware(doCheck bool) gin.HandlerFunc {
	//链接数据库
	//或者一些其他准备工作
	return func(c *gin.Context) {
		if doCheck {

			//存放具体的逻辑
			//是否登录的判断
			//if 是登录用户
			c.Next()
			//else
			//c.Abort()
		} else {
			c.Next()
		}
	}
}

func main() {
	r := gin.Default()

	r.Use(m1, m2, authMiddleware(true)) //全局注册中间件m1

	//string, handlers ...HandlerFunc) IRoutes {
	r.GET("/index", m1, indexHandler)
	r.GET("/shop", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg": "shop",
		})
	})
	r.GET("/user", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"msg": "user",
		})
	})

	r.Run(":8080")
}
