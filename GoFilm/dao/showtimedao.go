package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
)

// AddShowtime 添加场次（id 自增）
func AddShowtime(st *model.Showtime) error {
	sqlStr := "insert into showtimes(movie_id,starttime,hall_id,cinema_id,status,price) values(?,?,?,?,?,?)"
	result, err := utils.Db.Exec(sqlStr, st.MovieID, st.StartTime, st.HallID, st.CinemaID, st.Status, st.Price)
	if err != nil {
		fmt.Println("AddShowtime error:", err)
		return err
	}
	// 回填自增 ID
	id, _ := result.LastInsertId()
	st.ID = int(id)
	return nil
}

func GetShowtime() ([]*model.Showtime, error) {
	sqlStr := "select id,movie_id,starttime,hall_id,cinema_id,status,price from showtimes"
	rows, err := utils.Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Showtime

	for rows.Next() {
		st := &model.Showtime{}
		rows.Scan(&st.ID, &st.MovieID, &st.StartTime, &st.HallID, &st.CinemaID, &st.Status, &st.Price)
		//把影厅
		movie, _ := GetMovieById(st.MovieID)
		st.MovieName = movie.Title
		hall, _ := GetHallById(st.HallID)
		st.Hall = hall
		list = append(list, st)
	}

	return list, nil
}

// GetShowtimeById 根据场次 ID 查询
func GetShowtimeById(id int) (*model.Showtime, error) {
	sqlStr := "select id,movie_id,starttime,hall_id,cinema_id,status,price from showtimes where id = ?"
	st := &model.Showtime{}
	err := utils.Db.QueryRow(sqlStr, id).Scan(&st.ID, &st.MovieID, &st.StartTime, &st.HallID, &st.CinemaID, &st.Status, &st.Price)
	if err != nil {
		return nil, err
	}
	return st, nil
}

// GetShowtimesByCinemaId 根据影院 ID 查询所有场次
func GetShowtimesByCinemaId(cinemaID string) ([]*model.Showtime, error) {
	sqlStr := "select id,movie_id,starttime,hall_id,cinema_id,status,price from showtimes where cinema_id = ? order by starttime"
	rows, err := utils.Db.Query(sqlStr, cinemaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Showtime

	for rows.Next() {
		st := &model.Showtime{}
		rows.Scan(&st.ID, &st.MovieID, &st.StartTime, &st.HallID, &st.CinemaID, &st.Status, &st.Price)
		//把影厅
		movie, _ := GetMovieById(st.MovieID)
		st.MovieName = movie.Title
		hall, _ := GetHallById(st.HallID)
		st.Hall = hall
		list = append(list, st)
	}

	return list, nil
}

// GetShowtimesByHallId 根据影厅 ID 查询所有场次
func GetShowtimesByHallId(hallID string) ([]*model.Showtime, error) {
	sqlStr := "select id,movie_id,starttime,hall_id,cinema_id,status,price from showtimes where hall_id = ? order by starttime"
	rows, err := utils.Db.Query(sqlStr, hallID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Showtime
	for rows.Next() {
		st := &model.Showtime{}
		rows.Scan(&st.ID, &st.MovieID, &st.StartTime, &st.HallID, &st.CinemaID, &st.Status, &st.Price)
		hall, _ := GetHallById(st.HallID)
		st.Hall = hall
		list = append(list, st)
	}
	return list, nil
}

// GetShowtimesByMovieId 根据电影 ID 查询所有场次
func GetShowtimesByMovieId(movieID int) ([]*model.Showtime, error) {
	sqlStr := "select id,movie_id,starttime,hall_id,cinema_id,status,price from showtimes where moive_id = ? and status='上映中' order by starttime"
	rows, err := utils.Db.Query(sqlStr, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Showtime
	for rows.Next() {
		st := &model.Showtime{}
		rows.Scan(&st.ID, &st.MovieID, &st.StartTime, &st.HallID, &st.CinemaID, &st.Status, &st.Price)
		hall, _ := GetHallById(st.HallID)
		st.Hall = hall
		list = append(list, st)
	}
	return list, nil
}

// 根据电影id和影院id 查询所有场次
func GetShowtimeByMovieIdAndCinemaId(cinemaID string, movieId int) ([]*model.Showtime, error) {
	sqlStr := "select id,movie_id,starttime,hall_id,cinema_id,status,price from showtimes where cinema_id = ? and movie_id = ? order by starttime"
	rows, err := utils.Db.Query(sqlStr, cinemaID, movieId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Showtime

	for rows.Next() {
		st := &model.Showtime{}
		rows.Scan(&st.ID, &st.MovieID, &st.StartTime, &st.HallID, &st.CinemaID, &st.Status, &st.Price)
		//把影厅
		hall, _ := GetHallById(st.HallID)
		st.Hall = hall
		list = append(list, st)
	}

	return list, nil
}

// UpdateShowtime 更新场次（电影、时间、影厅、状态、价格）
func UpdateShowtime(st *model.Showtime) error {
	sqlStr := "update showtimes set  starttime=?, hall_id=?, status=?, price=? where id=?"
	_, err := utils.Db.Exec(sqlStr, st.StartTime, st.HallID, st.Status, st.Price, st.ID)
	if err != nil {
		fmt.Println("UpdateShowtime error:", err)
		return err
	}
	return nil
}

// DeleteShowtime 删除场次
func DeleteShowtime(id int) error {
	_, err := utils.Db.Exec("delete from showtimes where id = ?", id)
	if err != nil {
		//fmt.Println("DeleteShowtime error:", err)
		return err
	}
	return nil
}
