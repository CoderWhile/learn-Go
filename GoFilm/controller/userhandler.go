package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
	"html/template"
	"net/http"
)

// 处理用户注销
func Logout(w http.ResponseWriter, r *http.Request) {
	//获取cookie
	cookie, err := r.Cookie("user")
	if err != nil {
		fmt.Println(err)
	}
	if cookie != nil {

		//获取cookie的值
		cookieValue := cookie.Value
		//删除数据库中对应的Session
		dao.DeleteSession(cookieValue)
		//设置cookie失效
		cookie.MaxAge = -1
		//将修改之后的cookie发送给浏览器
		http.SetCookie(w, cookie)

	}
	//去首页
	FirstPage(w, r)
}

// 处理登录函数
func Login(w http.ResponseWriter, r *http.Request) {
	//判断是否已经登录
	//flag, _ := dao.IsLogin(r)
	//if flag {
	//	//已经登录
	//	//去首页
	//	GetPageBooksByPrice(w, r)
	//} else {
	//获取用户名和密码和身份信息
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	//一并处理用户和管理员登录
	//identity := r.PostFormValue("identity")
	user, _ := dao.CheckUserNamePasswordIdentity(username, password)
	if user.ID > 0 {
		//生成UUID作为Session的id
		uuid := utils.CreateUUID()
		//用户名密码正确
		//创建一个Session
		sess := &model.Session{
			SessionID:    uuid,
			UserName:     user.Username,
			UserID:       user.ID,
			UserIdentity: user.Identity,
		}
		//将Session保存到数据库
		dao.AddSession(sess)
		//创建一个Cookid,和Session关联
		cookie := http.Cookie{
			Name:     "user",
			Value:    uuid,
			HttpOnly: true,
		}
		//将Cookid发送给浏览器
		http.SetCookie(w, &cookie)
		//登录成功
		t := template.Must(template.ParseFiles("views/pages/user/login_success.html"))
		t.Execute(w, user)

	} else {
		t := template.Must(template.ParseFiles("views/pages/user/login.html"))
		t.Execute(w, "用户名或密码不正确")
	}
	//}

}

// 处理注册的函数
// 处理用户注册的函数
func Regist(w http.ResponseWriter, r *http.Request) {
	//获取用户名和密码
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	user, _ := dao.CheckUserName(username)
	fmt.Println(user)
	if user.ID > 0 {
		//用户名存在
		t := template.Must(template.ParseFiles("views/pages/user/regist.html"))
		t.Execute(w, "用户名已存在")

	} else {
		//将用户保存到数据中
		identity := "0"
		dao.SaveUser(username, password, identity)

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
