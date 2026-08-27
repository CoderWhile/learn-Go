package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	_, sess := dao.IsLogin(r)
	//
	tickets, _ := dao.GetTicketsByUserId(sess.UserID)
	var orders []model.OrderItem
	for _, tk := range tickets {
		st, _ := dao.GetShowtimeById(tk.ShowtimeID)
		if st == nil {
			continue
		}
		movie, _ := dao.GetMovieById(st.MovieID)
		cinema, _ := dao.GetCinemaById(st.CinemaID)
		hall, _ := dao.GetHallById(st.HallID)
		seat, _ := dao.GetSeatById(tk.SeatID)
		loc, _ := time.LoadLocation("Asia/Shanghai")
		now := time.Now().In(loc)
		status := "未放映"
		isPassed := false
		showTime, _ := time.ParseInLocation("2006-01-02T15:04", st.StartTime, loc)
		if !showTime.After(now) {
			status = "已放映"
			isPassed = true
		}

		item := model.OrderItem{
			TicketID:  tk.ID,
			ShowTime:  st.StartTime,
			Price:     st.Price,
			Status:    status,
			IsPassed:  isPassed,
			OrderTime: tk.CreatedAt,
		}
		if movie != nil {
			item.MovieTitle = movie.Title
			item.MovieImage = movie.ImagePath
		}
		if cinema != nil {
			item.CinemaName = cinema.Name
		}
		if hall != nil {
			item.HallName = hall.Name
		}
		if seat != nil {
			item.Row = seat.Row
			item.Col = seat.Col
			item.SeatLabel = seat.Row + strconv.Itoa(seat.Col)
			fmt.Println(item.SeatLabel)
		}
		orders = append(orders, item)
	}

	page := &model.MyOrderPage{
		Orders:   orders,
		Username: sess.UserName,
		IsLogin:  true,
	}

	t := template.Must(template.ParseFiles("views/pages/user/my_orders.html"))
	t.Execute(w, page)
}

// 退票
func RefundTicketHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, sess := dao.IsLogin(r)
	if sess == nil {
		w.Write([]byte("请先登录"))
		return
	}
	tid, _ := strconv.Atoi(r.PostFormValue("ticketId"))
	fmt.Println("tid", tid)
	ticket, _ := dao.GetTicketByID(tid)
	fmt.Printf("%+v\n", ticket)
	showtime, _ := dao.GetShowtimeById(ticket.ShowtimeID)
	fmt.Printf("%+v", showtime)
	layout := "2006-01-02T15:04"
	st, _ := time.Parse(layout, showtime.StartTime)
	elapsed := st.Sub(time.Now())
	//elapsed := time.Now().Sub(st)
	fmt.Println("elapsed", elapsed)
	if elapsed.Minutes() < 120 {
		w.Write([]byte("开场前两小时不需退票"))
	} else {

		err := dao.DeleteTicketByID(tid)
		if err != nil {
			w.Write([]byte("退票失败"))
			return
		}
		dao.UpdateBoxOffice(-showtime.Price, showtime.MovieID)
		w.Write([]byte("ok"))
	}
}

// 未支付订单页面（锁定中的票）
func UnpaidOrdersHandler(w http.ResponseWriter, r *http.Request) {
	_, sess := dao.IsLogin(r)
	if sess == nil {
		http.Redirect(w, r, "/pages/user/login.html", http.StatusFound)
		return
	}
	tickets, _ := dao.GetTicketsByUserId(sess.UserID)
	var orders []model.OrderItem
	var total float64
	for _, tk := range tickets {
		// 只展示锁定（未支付）的票
		if tk.Status != model.TicketStatusLocked {
			continue
		}
		st, _ := dao.GetShowtimeById(tk.ShowtimeID)
		if st == nil {
			continue
		}
		movie, _ := dao.GetMovieById(st.MovieID)
		cinema, _ := dao.GetCinemaById(st.CinemaID)
		hall, _ := dao.GetHallById(st.HallID)
		seat, _ := dao.GetSeatById(tk.SeatID)

		item := model.OrderItem{
			TicketID:  tk.ID,
			ShowTime:  st.StartTime,
			Price:     st.Price,
			OrderTime: tk.CreatedAt,
		}
		if movie != nil {
			item.MovieTitle = movie.Title
			item.MovieImage = movie.ImagePath
		}
		if cinema != nil {
			item.CinemaName = cinema.Name
		}
		if hall != nil {
			item.HallName = hall.Name
		}
		if seat != nil {
			item.Row = seat.Row
			item.Col = seat.Col
			item.SeatLabel = seat.Row + strconv.Itoa(seat.Col)
		}
		total += st.Price
		orders = append(orders, item)
	}

	page := &model.UnpaidOrderPage{
		Orders:   orders,
		Username: sess.UserName,
		IsLogin:  true,
		Total:    total,
	}
	t := template.Must(template.ParseFiles("views/pages/user/unpaid_orders.html"))
	t.Execute(w, page)
}

// 一键支付所有锁定票
func PayAllLockedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, sess := dao.IsLogin(r)
	if sess == nil {
		w.Write([]byte("请先登录"))
		return
	}
	// 先累计票房，再支付
	tickets, _ := dao.GetTicketsByUserId(sess.UserID)
	boxMap := make(map[int]float64)
	for _, tk := range tickets {
		if tk.Status != model.TicketStatusLocked {
			continue
		}
		st, _ := dao.GetShowtimeById(tk.ShowtimeID)
		if st != nil {
			boxMap[st.MovieID] += tk.Price
		}
	}
	if err := dao.PayLockedTickets(sess.UserID); err != nil {
		w.Write([]byte("支付失败"))
		return
	}
	for mid, amount := range boxMap {
		dao.UpdateBoxOffice(amount, mid)
	}
	w.Write([]byte("ok"))
}

// 取消锁定单张票
func CancelLockHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, sess := dao.IsLogin(r)
	if sess == nil {
		w.Write([]byte("请先登录"))
		return
	}
	tid, _ := strconv.Atoi(r.PostFormValue("ticketId"))
	if err := dao.DeleteLockedTicketByID(tid, sess.UserID); err != nil {
		w.Write([]byte("取消失败"))
		return
	}
	w.Write([]byte("ok"))
}
