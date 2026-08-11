package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
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
	movie, _ := dao.GetMovieById(int(imovieID))
	hallID := r.PostFormValue("hallId")
	startTime := r.PostFormValue("startTime")
	layout := "2006-01-02T15:04"
	st, _ := time.Parse(layout, startTime)

	price := r.PostFormValue("price")
	iprice, _ := strconv.ParseFloat(price, 64)
	status := r.PostFormValue("status")

	// 检查同一个影厅的时间冲突

	duration := movie.Duration
	newEnd := st.Add(time.Duration(duration) * time.Minute)
	existings, _ := dao.GetShowtimesByHallId(hallID)
	for _, es := range existings {
		if es.ID == 0 || es.Status == "已放映" {
			continue
		}
		esStart, err := time.Parse(layout, es.StartTime)
		if err != nil {
			continue
		}
		emovie, _ := dao.GetMovieById(es.MovieID)
		ed := 120
		if emovie != nil && emovie.Duration > 0 {
			ed = emovie.Duration
		}
		esEnd := esStart.Add(time.Duration(ed) * time.Minute)
		// 时间重叠判断: A.start < B.end && B.start < A.end
		if st.Before(esEnd) && esStart.Before(newEnd) {
			w.Write([]byte("该影厅在 " + es.StartTime + " 已有排片（电影：" + emovie.Title + "），时间冲突"))
			return
		}
	}
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

	// 检查同一个影厅的时间冲突
	movie, _ := dao.GetMovieById(int(imovieID))
	layout := "2006-01-02T15:04"
	st, _ := time.Parse(layout, StartTime) // ← 用 StartTime，不是 startTime
	duration := movie.Duration
	if duration <= 0 {
		duration = 120
	}
	newEnd := st.Add(time.Duration(duration) * time.Minute)
	existings, _ := dao.GetShowtimesByHallId(hallID)
	for _, es := range existings {
		if es.ID == int(sid) || es.Status == "已放映" { // ← 跳过自己 + 已放映
			continue
		}
		esStart, err := time.Parse(layout, es.StartTime)
		if err != nil {
			continue
		}
		emovie, _ := dao.GetMovieById(es.MovieID)
		ed := 120
		if emovie != nil && emovie.Duration > 0 {
			ed = emovie.Duration
		}
		esEnd := esStart.Add(time.Duration(ed) * time.Minute)
		// 时间重叠判断: A.start < B.end && B.start < A.end
		if st.Before(esEnd) && esStart.Before(newEnd) {
			w.Write([]byte("该影厅在 " + es.StartTime + " 已有排片（电影：" + emovie.Title + "），时间冲突"))
			return
		}
	}
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
	//更新所有该场次的票
	dao.UpdatePriceByShowtime(showtime.ID, iPrice)
	w.Write([]byte("ok"))

}

// 删除场次
func DeleteShowTimeHandler(w http.ResponseWriter, r *http.Request) {
	//
	sid := r.PostFormValue("showId")
	isid, _ := strconv.ParseInt(sid, 10, 64)
	//检查该场次有没有卖出票有就不删
	//tickets, _ := dao.GetTicketsByShowtimeId(int(isid))

	err := dao.DeleteShowtime(int(isid))
	if err == nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	} else {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("已经开始售票"))
	}
}

// 场次过期检查：每分钟将已过时间的"预售"场次改为"已放映"
func CheckStatusShowtime(ctx context.Context) {
	showtimes, _ := dao.GetShowtime()
	now := time.Now()
	layout := "2006-01-02T15:04"

	for _, s := range showtimes {
		// 已经是"已放映"的就跳过
		if s.Status == "已放映" {
			continue
		}
		st, err := time.Parse(layout, s.StartTime)
		if err != nil {
			continue
		}
		if st.Before(now) {
			s.Status = "已放映"
			fmt.Println("正在更新")
			dao.UpdateShowtime(s) // 直接用查出来的 s 更新，只改 status
		}
	}
}
