package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"html/template"
	"net/http"
)

// 查看影院
func CinemaHandler(w http.ResponseWriter, r *http.Request) {
	//加载所有影院信息
	cinemas, _ := dao.GetCinemas()
	pagecinema := &model.PageCianema{
		Cinemas:          cinemas,
		TotalCinemaCount: len(cinemas),
	}
	t := template.Must(template.ParseFiles("views/pages/cinema/cinema_list.html"))
	t.Execute(w, pagecinema)
}

// 添加电影院
func AddCinemaHandler(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	address := r.PostFormValue("address")
	intro := r.PostFormValue("intro")
	cinema := &model.Cinema{
		Name:    name,
		Address: address,
		Intro:   intro,
	}
	dao.AddCinema(cinema)
	t := template.Must(template.ParseFiles("views/pages/manager/cinema_add_success.html"))
	t.Execute(w, cinema)

	//
}
