package dao

import (
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
)

// 根据用户名和密码查询一条记录
func CheckUserNamePassword(username string, password string) (*model.User, error) {
	sqlStr := "select id,username,password,email from users where username = ? and password = ?"
	row := utils.Db.QueryRow(sqlStr, username, password)
	user := &model.User{}
	row.Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	return user, nil

}

// 验证用户名
func CheckUserName(username string) (*model.User, error) {
	sqlStr := "select id,username,password,email from users where username = ? "
	row := utils.Db.QueryRow(sqlStr, username)
	user := &model.User{}
	row.Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	return user, nil
}

// 保存User,插入信息
func SaveUser(username string, password string, email string) error {
	sqlStr := "insert into users(username,password,email) values(?,?,?)"
	_, err := utils.Db.Exec(sqlStr, username, password, email)
	if err != nil {
		return err
	}
	return nil
}
