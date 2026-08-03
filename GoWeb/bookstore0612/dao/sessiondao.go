package dao

import (
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
	"fmt"
	"net/http"
)

// AddSession向数据库中添加Session
func AddSession(sess *model.Session) error {
	sqlStr := `insert into sessions values (?,?,?)`
	_, err := utils.Db.Exec(sqlStr, sess.SessionID, sess.UserName, sess.UserID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// 删除
func DeleteSession(sessID string) error {
	sqlStr := `delete from sessions where session_id = ?`
	_, err := utils.Db.Exec(sqlStr, sessID)
	if err != nil {
		return err
	}
	return nil

}

// 根据sessionID获取session
func GetSession(sessID string) (*model.Session, error) {
	sqlStr := `select * from sessions where session_id = ?`
	inStmt, err := utils.Db.Prepare(sqlStr)
	if err != nil {
		return nil, err
	}
	row := inStmt.QueryRow(sessID)
	sess := &model.Session{}
	row.Scan(&sess.SessionID, &sess.UserName, &sess.UserID)
	return sess, nil
}

// 判断用户是否已经登录,true已经的登录
func IsLogin(r *http.Request) (bool, *model.Session) {
	//根据cookie的名字获取cookie
	cookie, _ := r.Cookie("user")
	if cookie != nil {
		cookieValue := cookie.Value
		session, _ := GetSession(cookieValue)
		if session.UserID > 0 {
			//已经登录
			return true, session
		}

	}
	return false, nil

}
