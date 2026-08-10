package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
	"time"
)

func InsertTicket(price float64, showtimeID, userID int, seatID int64) error {
	fmt.Println("插入一张票")
	sqlStr := `INSERT INTO tickets (price,showtime_id, user_id, seat_id, status, lock_time, created_at)
               VALUES (?,?, ?, ?, 'paid', ?, NOW())`
	_, err := utils.Db.Exec(sqlStr, price, showtimeID, userID, seatID, time.Now())
	if err != nil {

		fmt.Println(err)
		return err
	}
	return nil
}

// 获取该场次已经售卖或者锁定的座位id
func GetSoldSeatIdsByShowtime(showtimeID int) (map[int64]bool, error) {
	sqlStr := `SELECT seat_id FROM tickets WHERE showtime_id = ? AND status IN ('paid', 'locked')`
	rows, err := utils.Db.Query(sqlStr, showtimeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sold := make(map[int64]bool)
	for rows.Next() {
		var seatID int64
		rows.Scan(&seatID)
		sold[seatID] = true
	}
	return sold, nil
}

// GetTicketsByUserId 查询用户的购票记录
func GetTicketsByUserId(userID int) ([]*model.Ticket, error) {
	sqlStr := `SELECT id, showtime_id, user_id, seat_id, status, lock_time, created_at
               FROM tickets WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := utils.Db.Query(sqlStr, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Ticket
	for rows.Next() {
		t := &model.Ticket{}
		rows.Scan(&t.ID, &t.ShowtimeID, &t.UserID, &t.SeatID, &t.Status, &t.LockTime, &t.CreatedAt)
		list = append(list, t)
	}
	return list, nil
}
