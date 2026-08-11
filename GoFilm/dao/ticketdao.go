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

// 根据场次Id查找票
func GetTicketsByShowtimeId(stID int) ([]*model.Ticket, error) {
	sql := `select id,showtime_id,user_id,seat_id,status,lock_time,created_at from tickets where showtime_id=?`
	rows, err := utils.Db.Query(sql, stID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Ticket
	for rows.Next() {
		t := &model.Ticket{}
		rows.Scan(&t.ID, &t.ShowtimeID, &t.UserID, &t.SeatID, &t.Status, &t.LockTime, &t.CreatedAt)
	}
	return list, nil
}

// 根据座位id 删除票
func DeleteTicketBySeatId(stID int) error {
	sql := `delete * from tickets where seat_id=?`
	_, err := utils.Db.Exec(sql, stID)
	if err != nil {
		return err
	}
	return nil

}

// 根据用户票的id删除票
func DeleteTicketByID(id int) error {
	sql := `DELETE FROM tickets WHERE id=?`
	_, err := utils.Db.Exec(sql, id)
	if err != nil {
		return err
	}
	return nil
}

// 更新场次票价
func UpdatePriceByShowtime(stID int, price float64) error {
	sql := `update tickets set price=? where showtime_id=?`
	_, err := utils.Db.Exec(sql, price, stID)
	if err != nil {
		return err
	}
	return nil
}

func GetTicketByID(id int) (*model.Ticket, error) {
	sqlStr := `SELECT id,showtime_id,user_id,seat_id,status FROM tickets WHERE id = ? `
	row := utils.Db.QueryRow(sqlStr, id)
	t := &model.Ticket{}
	err := row.Scan(&t.ID, &t.ShowtimeID, &t.UserID, &t.SeatID, &t.Status)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return t, nil
}
