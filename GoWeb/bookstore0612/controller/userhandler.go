package controller

import (
	"GoWeb/bookstore0612/dao"
	"fmt"
	"html/template"
	"net/http"
)

//处理用户登录的函数

func Login(w http.ResponseWriter, r *http.Request) {
	//获取用户名和密码
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	user, _ := dao.CheckUserNamePassword(username, password)
	if user.ID > 0 {
		//用户名密码正确
		t := template.Must(template.ParseFiles("views/pages/user/login_success.html"))
		t.Execute(w, "")

	} else {
		t := template.Must(template.ParseFiles("views/pages/user/login.html"))
		t.Execute(w, "用户名或密码不正确")
	}

}

// 处理用户注册的函数
func Regist(w http.ResponseWriter, r *http.Request) {
	//获取用户名和密码
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	email := r.PostFormValue("email")
	user, _ := dao.CheckUserName(username)
	fmt.Println(user)
	if user.ID > 0 {
		//用户名存在
		t := template.Must(template.ParseFiles("views/pages/user/regist.html"))
		t.Execute(w, "用户名已存在")

	} else {
		//将用户保存到数据中
		dao.SaveUser(username, password, email)
		t := template.Must(template.ParseFiles("views/pages/user/regist_success.html"))
		t.Execute(w, "")
	}

}

// 通过发送ajax验证用户名是否可用
func CheckUserName(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	fmt.Println("传入的用户名")
	user, _ := dao.CheckUserName(username)
	if user.ID > 0 {
		//用户名存在
		w.Write([]byte("用户名已存在"))
	} else {
		//将用户保存到数据中
		w.Write([]byte("<font style='color:green'>用户名可用</font>"))

	}
}
