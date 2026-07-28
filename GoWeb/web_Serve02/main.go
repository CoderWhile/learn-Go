package main

import (
	"fmt"
	"net/http"
	"time"
)

type MyHandler struct{}

func (m *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "通过自己创建的处理器处理请求")
}
func main() {
	myHandler := MyHandler{}
	//创建Server结构
	//并详细配置服务器
	server := http.Server{
		Addr:        ":8080",
		Handler:     &myHandler,
		ReadTimeout: 10 * time.Second,
	}
	server.ListenAndServe()
}
