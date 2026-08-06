package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
)

// 向数据库中添加一个影院
func AddCinema(cinema *model.Cinema) error {
	id := utils.CreateUUID()
	sqlStr := "insert into cinemas(id,name,address,intro) values(?,?,?,?)"
	_, err := utils.Db.Exec(sqlStr, id, cinema.Name, cinema.Address, cinema.Intro)

	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println(cinema)
	return nil

}
