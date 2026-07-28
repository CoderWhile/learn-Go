package main

import (
	"GoWeb/web01_db/utils"
	"fmt"
)

type User struct {
	ID       int
	Username string
	Password string
	Email    string
}

// 添加：
func (user *User) AddUser() error {
	sqlStr := `INSERT INTO users(username,password,email) VALUES(?,?,?)`
	stmt, err := utils.Db.Prepare(sqlStr)
	if err != nil {
		fmt.Println("预编译出现异常", err)
		return err
	}

	_, err = stmt.Exec("admin", "123456", "admin@123")
	if err != nil {
		fmt.Println("执行出现异常", err)
		return err
	}
	return nil
}

func (user *User) AddUser2() error {
	sqlStr := `INSERT INTO users(username,password,email) VALUES(?,?,?)`
	_, err := utils.Db.Exec(sqlStr, "admin2", "333333", "dif@admin")

	if err != nil {
		fmt.Println("执行出现异常", err)
		return err
	}
	return nil
}

// 查询一条记录根据用户id
func (user *User) GetUserByID() (*User, error) {
	sqlStr := `SELECT id,username,password,email FROM users WHERE id=?`
	row := utils.Db.QueryRow(sqlStr, user.ID)
	var id int
	var username string
	var password string
	var email string
	err := row.Scan(&id, &username, &password, &email)
	if err != nil {

		return nil, err
	}
	u := &User{
		ID:       id,
		Username: username,
		Password: password,
		Email:    email,
	}
	return u, nil
}

// 查询多条
func (user *User) GetUsers() ([]*User, error) {
	sqlStr := `SELECT id,username,password,email FROM users`
	rows, err := utils.Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	//创建切片
	var users []*User
	for rows.Next() {
		var id int
		var username string
		var password string
		var email string
		err := rows.Scan(&id, &username, &password, &email)
		if err != nil {

			return nil, err
		}
		u := &User{
			ID:       id,
			Username: username,
			Password: password,
			Email:    email,
		}
		users = append(users, u)

	}
	return users, nil
}
