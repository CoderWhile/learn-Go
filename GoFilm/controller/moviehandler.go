package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"html/template"
	"net/http"
	"strconv"
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
		page.TotalMovieCount = len(movies)
	}
	page.Movies = movies
	t := template.Must(template.ParseFiles("views/index.html"))
	t.Execute(w, page)
}

func FirstPageManager(w http.ResponseWriter, r *http.Request) {
	flag, session := dao.IsLogin(r)
	page := &model.Page{}
	movies, _ := dao.GetMovies()
	if flag {
		page.IsLogin = true
		page.Username = session.UserName
		page.Movies = movies
		page.TotalMovieCount = len(movies)
	}
	page.Movies = movies
	t := template.Must(template.ParseFiles("views/pages/manager/manager.html"))
	t.Execute(w, page)
}

// 添加电影
func AddMovieHandler(w http.ResponseWriter, r *http.Request) {
	title := r.PostFormValue("title")
	genre := r.PostFormValue("genre")
	area := r.PostFormValue("area")
	intro := r.PostFormValue("intro")
	imagepath := r.PostFormValue("imagePath")
	rating := r.PostFormValue("rating")
	status := r.PostFormValue("status")
	duration := r.PostFormValue("duration")
	irating, _ := strconv.ParseFloat(rating, 64)
	iduration, _ := strconv.ParseInt(duration, 10, 0)
	movie := &model.Movie{
		Title:     title,
		Genre:     genre,
		Area:      area,
		Intro:     intro,
		ImagePath: imagepath,
		Rating:    irating,
		Status:    status,
		Duration:  int(iduration),
	}
	dao.AddMovie(movie)
	t := template.Must(template.ParseFiles("views/pages/manager/movie_add_success.html"))
	t.Execute(w, movie)
}

// 电影信息更新
func ToMovieUpdateHandler(w http.ResponseWriter, r *http.Request) {
	movieid := r.FormValue("movieId")
	imovieid, _ := strconv.ParseInt(movieid, 10, 64)

	movie, _ := dao.GetMovieById(int(imovieid))
	t := template.Must(template.ParseFiles("views/pages/movie/movie_edit.html"))
	t.Execute(w, movie)

}

func MovieUpdateHandler(w http.ResponseWriter, r *http.Request) {

}
