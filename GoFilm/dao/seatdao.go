package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
)

// AddSeats 批量添加座位（用于创建影厅时初始化）
func AddSeats(hallID string, seats []*model.Seat) error {
	if len(seats) == 0 {
		return nil
	}
	fmt.Println("准备添加座位")
	sqlStr := "insert into seats(hall_id, row, col, seattype, status) values(?,?,?,?,?)"
	for _, s := range seats {
		result, err := utils.Db.Exec(sqlStr, hallID, s.Row, s.Col, int(s.SeatType), int(s.Status))
		if err != nil {
			return fmt.Errorf("AddSeats row=%s col=%d: %w", s.Row, s.Col, err)
		}
		id, _ := result.LastInsertId()
		s.ID = id
	}
	return nil
}

// GetSeatsByHallId 根据影厅 ID 查询该影厅的所有座位（按排+列排序）
func GetSeatsByHallId(hallID string) ([]*model.Seat, error) {
	sqlStr := "select id,hall_id, row, col, seattype, status from seats where hall_id = ? order by row, col"
	rows, err := utils.Db.Query(sqlStr, hallID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []*model.Seat
	for rows.Next() {
		s := &model.Seat{}
		var seatType int
		var status int
		rows.Scan(&s.ID, &s.HallID, &s.Row, &s.Col, &seatType, &status)
		s.SeatType = model.SeatType(seatType)
		s.Status = model.SeatStatus(status)
		seats = append(seats, s)
	}
	return seats, nil
}

// GetSeatById 根据座位 ID 查询单个座位
func GetSeatById(id int64) (*model.Seat, error) {
	sqlStr := "select id,hall_id, row, col, seattype, status from seats where id = ?"
	s := &model.Seat{}
	var seatType int
	var status int
	err := utils.Db.QueryRow(sqlStr, id).Scan(&s.ID, &s.HallID, &s.Row, &s.Col, &seatType, &status)
	if err != nil {
		return nil, err
	}
	s.SeatType = model.SeatType(seatType)
	s.Status = model.SeatStatus(status)
	return s, nil
}

// UpdateSeat 更新单个座位的类型和状态（维修/恢复/改情侣座等）
func UpdateSeat(seat *model.Seat) error {
	sqlStr := "update seats set seattype=?, status=? where id=?"
	_, err := utils.Db.Exec(sqlStr, int(seat.SeatType), int(seat.Status), seat.ID)
	if err != nil {
		return fmt.Errorf("UpdateSeat id=%d: %w", seat.ID, err)
	}
	return nil
}

// UpdateSeatStatus 批量更新指定座位的状态（购票选座时用）
func UpdateSeatStatus(hallID string, row string, col int, status model.SeatStatus) error {
	sqlStr := "update seats set status=? where hall_id=? and row=? and col=?"
	_, err := utils.Db.Exec(sqlStr, int(status), hallID, row, col)
	if err != nil {
		return fmt.Errorf("UpdateSeatStatus hall=%s row=%s col=%d: %w", hallID, row, col, err)
	}
	return nil
}

// DeleteSeatsByHallId 删除某个影厅的所有座位（删除影厅时级联调用）
func DeleteSeatsByHallId(hallID string) error {
	_, err := utils.Db.Exec("delete from seats where hall_id = ?", hallID)
	if err != nil {
		return fmt.Errorf("DeleteSeatsByHallId hall=%s: %w", hallID, err)
	}
	return nil
}

// GetSeatCountByHallId 统计某影厅的座位总数
func GetSeatCountByHallId(hallID string) int {
	var count int
	utils.Db.QueryRow("select count(*) from seats where hall_id = ?", hallID).Scan(&count)
	return count
}
