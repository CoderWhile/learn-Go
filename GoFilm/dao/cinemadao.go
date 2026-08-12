package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
)

// 获取所有影院
func GetCinemas() ([]*model.Cinema, error) {
	sqlStr := `select id,name,address,intro from cinemas`
	utils.Db.Query(sqlStr)
	rows, err := utils.Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	var cinemas []*model.Cinema
	for rows.Next() {
		var cinema model.Cinema
		rows.Scan(&cinema.ID, &cinema.Name, &cinema.Address, &cinema.Intro)
		cinemas = append(cinemas, &cinema)
	}
	return cinemas, nil
}

// 向数据库中添加一个影院
func AddCinema(cinema *model.Cinema) error {
	id := utils.CreateUUID()
	sqlStr := "insert into cinemas(id,name,address,intro) values(?,?,?,?)"
	_, err := utils.Db.Exec(sqlStr, id, cinema.Name, cinema.Address, cinema.Intro)

	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(id, cinema.Name, cinema.Intro, cinema.Address)

	return nil

}

// 根据影院id获取一个影院
func GetCinemaById(id string) (*model.Cinema, error) {
	sqlStr := "select id,name,address,intro from cinemas where id=?"
	row := utils.Db.QueryRow(sqlStr, id)
	cinema := &model.Cinema{}
	err := row.Scan(&cinema.ID, &cinema.Name, &cinema.Address, &cinema.Intro)

	if err != nil {
		return nil, err
	}

	return cinema, nil

}

// 删除影院
func DeleteCinemaById(id string) error {
	sqlStr := "delete from cinemas where id=?"
	_, err := utils.Db.Exec(sqlStr, id)
	fmt.Println("进入删除")
	if err != nil {

		return err
	}
	return nil
}

// 更新影院信息
func UpdateCinema(cinema *model.Cinema) error {
	sqlStr := `update cinemas set name=?,address=?,intro=? where id=?`
	_, err := utils.Db.Exec(sqlStr, cinema.Name, cinema.Address, cinema.Intro, cinema.ID)
	if err != nil {
		return err
	}
	return nil
}
