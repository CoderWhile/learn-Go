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
		showTime, err := time.Parse("2006-01-02T15:04", st.StartTime)
		isPassed := err == nil && time.Now().After(showTime)
		status := "未放映"
		if isPassed {
			status = "已放映"
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
