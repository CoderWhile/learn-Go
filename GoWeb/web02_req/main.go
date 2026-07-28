package main

import (
	"encoding/json"
	"net/http"
)

// 创建处理器函数
func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintln(w, "你发送的请求地址是:", r.URL.Path)
	//fmt.Fprintln(w, "你发送的请求地址后的查询字符串是:", r.URL.RawQuery)
	//fmt.Fprintln(w, "请求头中所有的信息有:", r.Header)
	//fmt.Fprintln(w, "请求头中Accept-Encoding的信息有:", r.Header["Accept-Encoding"])
	//fmt.Fprintln(w, "请求头中Accept-Encoding的属性值是有:", r.Header.Get("Accept-Encoding"))
	//
	////获取请求体中内容长度
	////len := r.ContentLength
	////body := make([]byte, len)
	//////
	////r.Body.Read(body)
	//////在浏览器中 显示
	////fmt.Fprintln(w, "请求体中的 内容:", string(body))
	////解析表单,在调用r.Form之前必须执行该操作
	//r.ParseForm()
	////获取请求参数
	//fmt.Fprintln(w, "请求参数有:", r.Form)
	//fmt.Fprintln(w, "POST请求的form表单中的请求参数有", r.PostForm)
	////直接调用Formvalue方法和PostFormValue直接获取请求参数的值
	//fmt.Fprintln(w, "URL中的请求参数的值", r.FormValue("user"))
	//fmt.Fprintln(w, "Form表单中的请求参数的值", r.PostFormValue("username"))
	//w.Write([]byte("你的请求我已经收到了"))

	html := `<html>
		<head>
			<title>测试相应内容为网页</title>
			<meta charset="utf-8"/>
			</head>
	<body>
	我是以网页的形式响应过来的!
	</body>
		</html>`
	w.Write([]byte(html))

}

func testJsonRes(w http.ResponseWriter, r *http.Request) {
	//设置响应头中内容的类型
	w.Header().Set("Content-Type", "application/json")
	user := User{
		ID:       1,
		Username: "admin",
		Password: "123456",
	}
	json, _ := json.Marshal(user)
	w.Write(json)
}

func testRedire(w http.ResponseWriter, r *http.Request) {
	//设置响应头的Locatio属性
	w.Header().Set("Location", "https://www.baidu.com")
	//设置响应的状态码
	//只能调用一次，且必须再w.Write()之前调用
	w.WriteHeader(302)
}

func main() {
	http.HandleFunc("/hello", handler)
	http.HandleFunc("/testJson", testJsonRes)
	http.HandleFunc("/test", testRedire)
	http.ListenAndServe(":8080", nil)

}
