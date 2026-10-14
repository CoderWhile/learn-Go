package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func M1(c *gin.Context) {
	fmt.Println("进入中间件1--最外层")
	c.Next()
	fmt.Println("退出中间件1--最外层")
}
func M2(c *gin.Context) {
	fmt.Println("  进入中间件2--中间层")
	c.Next()
	fmt.Println("  退出 中间件2 ── 中间层")
}
func M3(c *gin.Context) {
	fmt.Println("    进入 中间件3 ── 最内层")
	c.Next()
	fmt.Println("    退出 中间件3 ── 最内层")
}
func main() {
	r := gin.Default()
	r.Use(M1, M2, M3)
	r.GET("/hello", func(c *gin.Context) {
		fmt.Println("		执行业务Handler")
		c.JSON(200, gin.H{"msg": "hello"})
	})

	r.Run(":8080")
}
