package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"crypto/md5"
	"encoding/hex"
)

// 根据用户名和密码查询一条记录
func CheckUserNamePasswordIdentity(username string, password string) (*model.User, error) {
	sqlStr := "select id,username,password,identity from users where username = ? and password = ?"
	row := utils.Db.QueryRow(sqlStr, username, password)
	user := &model.User{}
	row.Scan(&user.ID, &user.Username, &user.Password, &user.Identity)
	return user, nil

}

// 验证用户名
func CheckUserName(username string) (*model.User, error) {
	sqlStr := "select id,username,password,identity from users where username = ? "
	row := utils.Db.QueryRow(sqlStr, username)
	user := &model.User{}
	row.Scan(&user.ID, &user.Username, &user.Password, &user.Identity)
	return user, nil
}

// 保存User,插入信息
func SaveUser(username string, password string, identity string) error {
	sqlStr := "insert into users(username,password,identity) values(?,?,?)"
	_, err := utils.Db.Exec(sqlStr, username, password, identity)
	if err != nil {
		return err
	}
	return nil
}

// MD5Hash 计算字符串的 MD5 哈希值
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
