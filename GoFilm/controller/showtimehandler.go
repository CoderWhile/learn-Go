package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// 去场次添加页面
func ToAddShowTimeHandler(w http.ResponseWriter, r *http.Request) {
	//获取当前影院的影厅和所有电影
	movies, _ := dao.GetMovies()
	//获得影院id
	cinemaID := r.FormValue("cinemaId")
	fmt.Println("影院id是", cinemaID)
	//根据影院查询影厅
	halls, _ := dao.GetHallsByCinemaId(cinemaID)
	pageshowtimeinfo := &model.PageShowtimeInfo{}
	pageshowtimeinfo.CinemaID = cinemaID
	pageshowtimeinfo.Movies = movies
	pageshowtimeinfo.Halls = halls
	t := template.Must(template.ParseFiles("views/pages/showtime/show_add.html"))
	t.Execute(w, pageshowtimeinfo)
}
func AddShowTimeHandler(w http.ResponseWriter, r *http.Request) {
	cinemaID := r.PostFormValue("cinemaId")
	movieID := r.PostFormValue("movieId")
	imovieID, _ := strconv.ParseInt(movieID, 10, 64)

	hallID := r.PostFormValue("hallId")
	startTime := r.PostFormValue("startTime")
	price := r.PostFormValue("price")
	iprice, _ := strconv.ParseFloat(price, 64)
	status := r.PostFormValue("status")

	showtime := &model.Showtime{

		CinemaID:  cinemaID,
		MovieID:   int(imovieID),
		StartTime: startTime,
		HallID:    hallID,
		Status:    status,
		Price:     iprice,
	}
	dao.AddShowtime(showtime)

	t := template.Must(template.ParseFiles("views/pages/showtime/show_add_success.html"))
	t.Execute(w, showtime)

}

// 场次更新
func UpdateShowTimeHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PostFormValue("showId")
	sid, _ := strconv.ParseInt(id, 10, 64)
	movieID := r.PostFormValue("movieId")
	imovieID, _ := strconv.ParseInt(movieID, 10, 64)
	hallID := r.PostFormValue("hallId")
	StartTime := r.PostFormValue("showTime")
	Price := r.PostFormValue("price")
	iPrice, _ := strconv.ParseFloat(Price, 64)
	Status := r.PostFormValue("status")
	oldshowtime, _ := dao.GetShowtimeById(int(sid))
	showtime := &model.Showtime{
		ID:        int(sid),
		MovieID:   int(imovieID),
		StartTime: StartTime,
		HallID:    hallID,
		CinemaID:  oldshowtime.CinemaID,
		Status:    Status,
		Price:     iPrice,
	}
	dao.UpdateShowtime(showtime)

}
