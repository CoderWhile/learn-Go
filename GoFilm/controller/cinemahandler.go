package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"fmt"
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

// 影院详细信息展示
func CinemaInfoHandler(w http.ResponseWriter, r *http.Request) {
	pagecinemainfo := &model.PageCinemaInfo{}
	//获取影院信息
	cinemaid := r.PostFormValue("cinemaId")
	cinema, err := dao.GetCinemaById(cinemaid)
	if err != nil {
		fmt.Println("根据影院id获取影院错误：", err)
	}
	pagecinemainfo.Cinema = cinema
	//获取该影院的影厅
	halls, err := dao.GetHallsByCinemaId(cinemaid)
	if err != nil {
		fmt.Println("根据影院id获取影厅错误")
	}
	pagecinemainfo.Halls = halls
	//获取该影院的场次
	showtimes, err := dao.GetShowtimesByCinemaId(cinemaid)
	pagecinemainfo.Showtimes = showtimes
	//解析模板
	t := template.Must(template.ParseFiles("views/pages/cinema/cinema_detail.html"))
	t.Execute(w, pagecinemainfo)
}
