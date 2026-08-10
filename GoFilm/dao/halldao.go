package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
	"strings"
)

// AddHall 向指定影院添加影厅（自动生成 UUID，Version 初始为 1）
func AddHall(cinemaID string, hall *model.Hall) error {
	hall.Version = 0
	sqlStr := "insert into halls(id,cinema_id,name,totalrows,totalcols,Version) values(?,?,?,?,?,?)"
	_, err := utils.Db.Exec(sqlStr, hall.ID, cinemaID, hall.Name, hall.TotalRows, hall.TotalCols, hall.Version)
	if err != nil {
		fmt.Println("AddHall error:", err)
		return err
	}
	return nil
}

// GetHallById 根据影厅 ID 查询单个影厅
func GetHallById(id string) (*model.Hall, error) {
	sqlStr := "select id,name,totalrows,totalcols,Version,cinema_id from halls where id = ?"
	hall := &model.Hall{}
	err := utils.Db.QueryRow(sqlStr, id).Scan(&hall.ID, &hall.Name, &hall.TotalRows, &hall.TotalCols, &hall.Version, &hall.CinemaID)
	if err != nil {
		return nil, err
	}
	return hall, nil
}

// GetHallsByCinemaId 根据影院 ID 查询该影院下所有影厅
func GetHallsByCinemaId(cinemaID string) ([]*model.Hall, error) {
	sqlStr := "select id,name,totalrows,totalcols,Version from halls where cinema_id = ?"
	rows, err := utils.Db.Query(sqlStr, cinemaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var halls []*model.Hall
	for rows.Next() {
		hall := &model.Hall{}
		rows.Scan(&hall.ID, &hall.Name, &hall.TotalRows, &hall.TotalCols, &hall.Version)
		halls = append(halls, hall)
	}
	return halls, nil
}

// UpdateHall 更新影厅名称、排数、座位数（乐观锁：Version 自增）
func UpdateHall(hall *model.Hall) error {
	sqlStr := "update halls set name=?, totalrows=?, totalcols=?, Version=Version+1 where id=? and Version=?"
	result, err := utils.Db.Exec(sqlStr, hall.Name, hall.TotalRows, hall.TotalCols, hall.ID, hall.Version)
	if err != nil {
		fmt.Println("UpdateHall error:", err)
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("更新失败：数据已被他人修改，请刷新后重试")
	}
	hall.Version++ // 同步内存中的版本号
	return nil
}

// DeleteHall 删除影厅（先删关联座位，再删影厅本身）
func DeleteHall(id string) error {
	_, err := utils.Db.Exec("delete from seats where hall_id = ?", id)
	if err != nil {
		fmt.Println("DeleteHall seats error:", err)
		return err
	}
	_, err = utils.Db.Exec("delete from halls where id = ?", id)
	if err != nil {
		fmt.Println("DeleteHall error:", err)
		return err
	}
	return nil
}

func rowToInt(row string) int {
	var n int
	for _, c := range strings.ToUpper(row) {
		n = n*26 + int(c-'A'+1)
	}
	return n
}
