package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
)

func GetAllCategories() ([]*model.Category, error) {
	rows, err := utils.Db.Query("SELECT id, name FROM categories ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*model.Category
	for rows.Next() {
		c := &model.Category{}
		rows.Scan(&c.ID, &c.Name)
		list = append(list, c)
	}
	return list, nil
}

func GetCategoryById(id int) (*model.Category, error) {
	fmt.Println(id)
	row := utils.Db.QueryRow("SELECT id, name FROM categories WHERE id=?", id)
	c := &model.Category{}
	err := row.Scan(&c.ID, &c.Name)

	if err != nil {
		return c, err
	}

	return c, nil
}
