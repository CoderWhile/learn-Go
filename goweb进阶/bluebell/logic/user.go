package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"bluebell/pkg/snowflake"
)

//存放业务逻辑的代码

func SignUp(p *models.ParamSignUp) (err error) {
	//1判断用户存不存在

	if err := mysql.CheckUserExist(p.Username); err != nil {
		return err
	}

	//2生成UID

	userID := snowflake.GenID()
	user := &models.User{
		UserID:   userID,
		Username: p.Username,
		Password: p.Password,
	}
	//构造一个user实例
	//3.密码加密

	//4保存数据库里
	return mysql.InsertUser(user)
}

func Login(p *models.ParamLogin) error {
	user := &models.User{
		Username: p.Username,
		Password: p.Password,
	}
	return mysql.Login(user)
}
