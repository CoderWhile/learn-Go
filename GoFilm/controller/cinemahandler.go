package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"net/http"
)

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
	//直接去查看该电影院的页面
	//
}
