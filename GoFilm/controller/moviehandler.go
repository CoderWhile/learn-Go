package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func FirstPage(w http.ResponseWriter, r *http.Request) {
	searchword := r.FormValue("keyword")
	region := r.FormValue("region")
	cid := r.FormValue("tag")
	icid, _ := strconv.Atoi(cid)
	//fmt.Println("icid:", icid)
	category := &model.Category{
		ID:   icid,
		Name: "",
	}
	category, _ = dao.GetCategoryById(icid)
	fmt.Println(region, "cid:", cid)
	fmt.Printf("%+v\n", category)
	flag, session := dao.IsLogin(r)
	page := &model.Page{CategoryID: icid, Keyword: searchword, Region: region}

	// 加载分类和标签（始终显示在页面上）
	categories, _ := dao.GetAllCategories()
	//tags, _ := dao.GetAllTags()
	page.Categories = categories
	//page.Tags = tags

	// 按条件查询电影
	var movies []*model.Movie
	if searchword != "" {
		movies, _ = dao.GetMoviesByWord(searchword)
	} else if region != "" || icid != 0 {
		movies, _ = dao.GetMoviesByReigonAndTag(region, category.Name)
	} else {
		movies, _ = dao.GetMovies()
	}

	if flag {
		page.IsLogin = true
		page.Username = session.UserName
	}
	page.Movies = movies
	page.TotalMovieCount = len(movies)

	// 票房排行榜（始终显示）
	boxOffice, _ := dao.GetMovieByBoxoffice()
	for i := range boxOffice {
		boxOffice[i].Rank = i + 1
	}
	page.BoxOfficeMovies = boxOffice

	t := template.Must(template.ParseFiles("views/index.html"))
	t.Execute(w, page)
}

func FirstPageManager(w http.ResponseWriter, r *http.Request) {
	searchword := r.FormValue("keyword")
	region := r.FormValue("region")
	cid := r.FormValue("tag")
	icid, _ := strconv.Atoi(cid)
	//fmt.Println("icid:", icid)
	category := &model.Category{
		ID:   icid,
		Name: "",
	}
	category, _ = dao.GetCategoryById(icid)
	flag, session := dao.IsLogin(r)
	page := &model.Page{CategoryID: icid, Keyword: searchword, Region: region}

	// 加载分类和标签（始终显示在页面上）
	categories, _ := dao.GetAllCategories()
	//tags, _ := dao.GetAllTags()
	page.Categories = categories
	//page.Tags = tags

	// 按条件查询电影
	var movies []*model.Movie
	if searchword != "" {
		movies, _ = dao.GetMoviesByWord(searchword)
	} else if region != "" || icid != 0 {
		movies, _ = dao.GetMoviesByReigonAndTag(region, category.Name)
	} else {
		movies, _ = dao.GetMovies()
	}
	if flag {
		page.IsLogin = true
		page.Username = session.UserName
	}
	page.Movies = movies
	page.TotalMovieCount = len(movies)
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
	t := template.Must(template.ParseFiles("views/pages/movie/movie_add_success.html"))
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
	movieid := r.PostFormValue("id")
	imovieid, _ := strconv.ParseInt(movieid, 10, 64)
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
		ID:        int(imovieid),
		Title:     title,
		Genre:     genre,
		Area:      area,
		Intro:     intro,
		ImagePath: imagepath,
		Rating:    irating,
		Status:    status,
		Duration:  int(iduration),
	}
	dao.UpdateMovie(movie)
	//电影修改成功
	FirstPageManager(w, r)
}

// 删除电影
func DeleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	//获取要删除的电影的id
	movieid := r.PostFormValue("movieId")
	imovieid, _ := strconv.ParseInt(movieid, 10, 64)

	err := dao.DeleteMovieById(int(imovieid))
	if err != nil {
		fmt.Println(err)
	}

	FirstPageManager(w, r)
}

// 进入电影详情界面
func MovieDetailHandler(w http.ResponseWriter, r *http.Request) {
	//调用IsLogin函数
	movieID := r.FormValue("movieId")
	imovieID, _ := strconv.ParseInt(movieID, 10, 64)
	flag, session := dao.IsLogin(r)
	page := &model.MovieDetail{}
	movie, _ := dao.GetMovieById(int(imovieID))
	var showtimeGroups []*model.ShowtimeGroup

	cinemas, _ := dao.GetCinemas()
	for _, cinema := range cinemas {
		//根据电影id和影院Id查询对应场次
		//当前电影院下当前电影的场次
		showtimes, _ := dao.GetShowtimeByMovieIdAndCinemaId(cinema.ID, int(imovieID))
		showtimeGroup := &model.ShowtimeGroup{
			CinemaName: cinema.Name,
			CinemaAddr: cinema.Address,
			Showtimes:  showtimes,
		}

		showtimeGroups = append(showtimeGroups, showtimeGroup)
	}
	if flag {
		//已经登录了,设置page中的IsLogin和Username
		page.IsLogin = true
		page.Username = session.UserName

	}
	if movie.Status == "下架" {
		page.Isdelist = false
	} else {
		page.Isdelist = true
	}
	page.Movie = movie
	page.ShowtimeGroups = showtimeGroups

	t := template.Must(template.ParseFiles("views/pages/movie/movie_detail.html"))
	t.Execute(w, page)
}
