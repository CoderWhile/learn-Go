package main

import (
	"fmt"
	"net/http"
)

// 创建处理器函数,参数固定
// http.ResponseWriter响应写入器,是一个接口,用来构建并发送HTTP
//响应给客户端,
//*http.Request:请求对象指针,封装了客户端发来的所有信息,URL.Methond,Header,Body,Cookie

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello!", r.URL.Path)
}

func main() {
	//注册路由,将URL路径与handler函数绑定
	//handleFunc内部将传入的函数包装称一个实现了http.Hanlder接口的适配器
	http.HandleFunc("/", handler) //
	//启动HTTP服务器并开始监听
	//第二个参数是http.Hanlder类型的路由器,传nil表示用全局默认的http.DefaultServeMux即上面的多路
	//阻塞调用,
	http.ListenAndServe(":8080", nil)
}
