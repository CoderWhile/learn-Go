package main

import (
	"database/sql"
	"fmt"
)

var db *sql.DB

func initDB(maxConn, MaxIdle int) (err error) {
	// 构建连接字符串 (DSN)
	// 格式: 用户名:密码@tcp(地址:端口)/数据库名
	// 如果你没改过配置，通常地址是 127.0.0.1，端口是 3306
	// 这里的 'mysql' 是系统自带的默认库，用来测试连接最方便
	dsn := "root:bjy12345@tcp(127.0.0.1:3306)/chatroom"

	fmt.Println("正在尝试连接数据库...")

	// 打开数据库连接
	//不能写冒号，会导致空指针引用
	db, err = sql.Open("mysql", dsn) //不会校验用户名和密码是否正确
	if err != nil {                  //dsn格式不正确报错
		return err
	}

	// 确保程序结束时关闭连接
	//defer db.Close()
	//尝试与数据库建立连接
	// Ping 一下数据库，验证连接是否真的通了
	err = db.Ping()
	if err != nil {
		return err
	}
	//设置数据库连接池的最大连接数
	db.SetMaxOpenConns(10)
	//最大空闲连接数
	db.SetMaxIdleConns(5)
	return nil

}
