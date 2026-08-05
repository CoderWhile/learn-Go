package main

import (
	"GoFilm/controller"
	"net/http"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	//设置处理静态资源
	//直接去页面
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("views/static/"))))
	http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.Dir("views/pages/"))))
	http.HandleFunc("/main", controller.FirstPage)

	//去登陆
	http.HandleFunc("/login", controller.Login)

	//去注册
	http.HandleFunc("/regist", controller.Regist)

	http.ListenAndServe(":8080", nil)
}
