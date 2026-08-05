package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"html/template"
	"net/http"
)

func FirstPage(w http.ResponseWriter, r *http.Request) {
	//调用IsLogin函数
	flag, session := dao.IsLogin(r)
	page := &model.Page{}
	movies, _ := dao.GetMovies()
	if flag {
		//已经登录了,设置page中的IsLogin和Username
		page.IsLogin = true
		page.Username = session.UserName
		page.Movies = movies
	}
	page.Movies = movies
	t := template.Must(template.ParseFiles("views/index.html"))
	t.Execute(w, page)
}
