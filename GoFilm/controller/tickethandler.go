package controller

import (
	"GoFilm/dao"
	"GoFilm/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

//// TicketSelectHandler 渲染选座购票页
//func TicketSelectHandler(w http.ResponseWriter, r *http.Request) {
//	sid := r.URL.Query().Get("showtimeId")
//	id, err := strconv.Atoi(sid)
//	if err != nil || id <= 0 {
//		http.Error(w, "参数错误", http.StatusBadRequest)
//		return
//	}
//
//	st, _ := dao.GetShowtimeById(id)
//	hall, _ := dao.GetHallById(st.HallID)
//	seats, _ := dao.GetSeatsByHallId(st.HallID)
//	soldIDs, _ := dao.GetSoldSeatIdsByShowtime(id)
//
//	flag, sess := dao.IsLogin(r)
//	page := &model.TicketPage{
//		Showtime:    st,
//		Hall:        hall,
//		Seats:       seats,
//		SoldSeatIDs: soldIDs,
//	}
//	if flag {
//		page.IsLogin = true
//		page.Username = sess.UserName
//	}
//
//	t := template.Must(template.ParseFiles("views/pages/ticket/seat_select.html"))
//	t.Execute(w, page)
//}

// GetTicketDataJSON 选座页 Ajax 数据接口
func GetTicketDataJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sid := r.URL.Query().Get("showtimeId")
	id, err := strconv.Atoi(sid)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "参数错误"})
		return
	}

	st, err := dao.GetShowtimeById(id)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "场次不存在"})
		return
	}

	hall, _ := dao.GetHallById(st.HallID)
	//获取该影厅的座位
	seats, _ := dao.GetSeatsByHallId(st.HallID)
	fmt.Printf("%+v\n", seats)
	//获取已经售卖的票的id
	soldIDs, _ := dao.GetSoldSeatIdsByShowtime(id)
	movie, _ := dao.GetMovieById(st.MovieID)
	cinema, _ := dao.GetCinemaById(st.CinemaID)

	type SeatVO struct {
		Row      string `json:"row"`
		Col      int    `json:"col"`
		SeatType int    `json:"seattype"`
		Status   int    `json:"status"`
		Sold     bool   `json:"sold"`
	}
	seatVOs := make([]SeatVO, 0, len(seats))
	//遍历每一个座位
	for _, s := range seats {
		seatVOs = append(seatVOs, SeatVO{
			Row: s.Row, Col: s.Col,
			SeatType: int(s.SeatType),
			Status:   int(s.Status),
			Sold:     soldIDs[s.ID],
		})
	}

	resp := map[string]interface{}{
		"id":         st.ID,
		"movieTitle": movie.Title,
		"movieImage": movie.ImagePath,
		"cinemaName": cinema.Name,
		"hallName":   hall.Name,
		"showTime":   st.StartTime,
		"price":      st.Price,
		"status":     st.Status,
		"seats":      seatVOs,
	}
	json.NewEncoder(w).Encode(resp)
}

// BuyTicketHandler 确认购票
func BuyTicketHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, sess := dao.IsLogin(r)
	if sess == nil {
		w.Write([]byte("请先登录"))
		return
	}

	sid := r.PostFormValue("showtimeId")
	seatsJSON := r.PostFormValue("seats")
	showtimeID, _ := strconv.Atoi(sid)

	type SeatInput struct {
		Row string `json:"row"`
		Col int    `json:"col"`
	}
	var selected []SeatInput
	json.Unmarshal([]byte(seatsJSON), &selected)

	if len(selected) == 0 {
		w.Write([]byte("请选择座位"))
		return
	}

	//查场次
	st, _ := dao.GetShowtimeById(showtimeID)
	//查影厅
	hall, _ := dao.GetHallById(st.HallID)
	//用影厅查作为
	hall.Seats, _ = dao.GetSeatsByHallId(st.HallID)
	hall.BuildMatrix()

	var seatIDs []int64
	for _, si := range selected {
		seat := hall.GetSeat(si.Row, si.Col)
		if seat == nil {
			w.Write([]byte("座位不存在：" + si.Row + strconv.Itoa(si.Col)))
			return
		}
		if seat.Status == 1 {
			w.Write([]byte("座位已维修：" + si.Row + strconv.Itoa(si.Col)))
			return
		}
		seatIDs = append(seatIDs, seat.ID)
	}

	//开启事务
	tx, err := utils.Db.Begin()
	if err != nil {
		w.Write([]byte("系统繁忙"))
		return
	}

	for _, seatID := range seatIDs {
		// 先尝试把已有的锁直接升级为 paid
		result, _ := tx.Exec(
			"UPDATE tickets SET status='paid', lock_time=NULL WHERE showtime_id=? AND seat_id=? AND user_id=? AND status='locked'",
			showtimeID, seatID, sess.UserID,
		)
		affected, _ := result.RowsAffected()
		if affected > 0 {
			continue
		}

		// 没锁就直接插入（UNIQUE 约束防并发）
		_, err = tx.Exec(
			"INSERT INTO tickets (price, showtime_id, user_id, seat_id, status, lock_time, created_at) VALUES (?, ?, ?, ?, 'paid', NULL, NOW())",
			st.Price, showtimeID, sess.UserID, seatID,
		)
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "Duplicate") {
				w.Write([]byte("部分座位已被他人抢先购买，请重新选座"))
			} else {
				w.Write([]byte("购票失败"))
			}
			return
		}
	}

	tx.Commit()
	w.Write([]byte("ok"))
}

// LockSeatsHandler 锁定座位（5分钟），防止其他用户选购
func LockSeatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, sess := dao.IsLogin(r)
	if sess == nil {
		w.Write([]byte("请先登录"))
		return
	}

	sid := r.PostFormValue("showtimeId")
	seatsJSON := r.PostFormValue("seats")
	showtimeID, _ := strconv.Atoi(sid)

	type SeatInput struct {
		Row string `json:"row"`
		Col int    `json:"col"`
	}
	var selected []SeatInput
	json.Unmarshal([]byte(seatsJSON), &selected)
	if len(selected) == 0 {
		w.Write([]byte("ok"))
		return
	}

	st, _ := dao.GetShowtimeById(showtimeID)
	hall, _ := dao.GetHallById(st.HallID)
	hall.Seats, _ = dao.GetSeatsByHallId(st.HallID)
	price := st.Price
	hall.BuildMatrix()

	//先把所有之前该用户的记录删掉
	utils.Db.Exec("DELETE FROM tickets WHERE showtime_id=? AND user_id=? AND status='locked'", showtimeID, sess.UserID)

	tx, _ := utils.Db.Begin()
	for _, si := range selected {
		seat := hall.GetSeat(si.Row, si.Col)
		if seat == nil {
			tx.Rollback()
			w.Write([]byte("座位不存在"))
			return
		}
		_, err := tx.Exec(
			"INSERT INTO tickets (price,showtime_id, user_id, seat_id, status, lock_time, created_at) VALUES (?,?, ?, ?, 'locked', DATE_ADD(NOW(), INTERVAL 5 MINUTE), NOW())",
			price, showtimeID, sess.UserID, seat.ID,
		)
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "UNIQUE") {
				w.Write([]byte("busy"))
			} else {
				w.Write([]byte("锁定失败"))
			}
			return
		}
	}
	tx.Commit()
	w.Write([]byte("ok"))
}

//如何实现电影名称的搜索功能,go语言中有什么和字符串匹配相关的函数

// 倒计时结束释放锁定
func UnlockSeatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, sess := dao.IsLogin(r)
	if sess == nil {
		w.Write([]byte("ok"))
		return
	}
	sid := r.PostFormValue("showtimeId")
	showtimeID, _ := strconv.Atoi(sid)
	utils.Db.Exec("DELETE FROM tickets WHERE showtime_id=? AND user_id=? AND status='locked'", showtimeID, sess.UserID)
	w.Write([]byte("ok"))
}
