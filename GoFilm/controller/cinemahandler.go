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
	t := template.Must(template.ParseFiles("views/pages/cinema/cinema_add_success.html"))
	t.Execute(w, cinema)

	//
}
func multiply(a, b int) int {
	return a * b
}

var funcMap = template.FuncMap{
	"multiply": multiply,
}

// 影院详细信息展示
func CinemaInfoHandler(w http.ResponseWriter, r *http.Request) {
	pagecinemainfo := &model.PageCinemaInfo{}
	//获取影院信息
	cinemaid := r.FormValue("cinemaId")

	cinema := &model.Cinema{}
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

	//// 必须：先Funcs注册函数 → 再解析模板！
	//tpl, err := template.New("").
	//	Funcs(funcMap).
	//	ParseFiles("views/pages/cinema/cinema_detail.html")
	//if err != nil {
	//	fmt.Println("模板解析失败：", err)
	//	http.Error(w, "页面模板加载失败", http.StatusInternalServerError)
	//	return
	//}
	//
	//// 执行渲染
	//err = tpl.Execute(w, pagecinemainfo)
	//if err != nil {
	//	fmt.Println("模板渲染失败：", err)
	//}

	//解析模板
	t := template.Must(template.ParseFiles("views/pages/cinema/cinema_detail.html"))
	t.Execute(w, pagecinemainfo)
}

// 删除影院
func DeleteCinemaHandler(w http.ResponseWriter, r *http.Request) {
	cinemaid := r.FormValue("cinemaId")

	err := dao.DeleteCinemaById(cinemaid)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("仍有放映任务"))
	} else {
		w.Write([]byte("ok"))
	}
	
}

// 用户端影院列表
func UserCinemaListHandler(w http.ResponseWriter, r *http.Request) {
	flag, session := dao.IsLogin(r)
	cinemaList := &model.CinemaList{}
	if flag {
		cinemaList.Username = session.UserName
		cinemaList.IsLogin = true
	}
	cinema, _ := dao.GetCinemas()
	cinemaList.Cinemas = cinema
	cinemaList.TotalCinemaCount = len(cinema)
	t := template.Must(template.ParseFiles("views/pages/cinema/cinema_list_user.html"))
	t.Execute(w, cinemaList)
}

// 用户影院排片页
func CinemaShowsHandler(w http.ResponseWriter, r *http.Request) {
	//得到电影院id
	cinemashows := &model.CinemaShows{}
	cinemaid := r.FormValue("cinemaId")
	cinema, _ := dao.GetCinemaById(cinemaid)
	movies, _ := dao.GetMovies()
	var moviegroups []*model.MovieGroup
	cinemashows.Cinema = cinema
	for _, movie := range movies {
		moviegroup := &model.MovieGroup{}
		moviegroup.MovieTitle = movie.Title
		moviegroup.MovieGenre = movie.Genre
		showtimes, _ := dao.GetShowtimeByMovieIdAndCinemaId(cinemaid, movie.ID)
		if showtimes != nil {
			moviegroup.IsShow = true
		} else {
			moviegroup.IsShow = false
		}
		moviegroup.Showtimes = showtimes

		moviegroups = append(moviegroups, moviegroup)
	}
	cinemashows.MovieGroups = moviegroups
	flag, session := dao.IsLogin(r)
	if flag {
		cinemashows.Username = session.UserName
		cinemashows.IsLogin = true
	}
	t := template.Must(template.ParseFiles("views/pages/cinema/cinema_shows.html"))
	t.Execute(w, cinemashows)

}
