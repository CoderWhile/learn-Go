package utils

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Db  *sql.DB
	err error
)

func init() {
	Db, err = sql.Open("mysql", "root:bjy12345@tcp(127.0.0.1:3306)/theater_data")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("正在尝试连接数据库...")

	if err != nil { //dsn格式不正确报错
		fmt.Println(err)
	}

	// 确保程序结束时关闭连接
	//defer db.Close()
	//尝试与数据库建立连接
	// Ping 一下数据库，验证连接是否真的通了
	err = Db.Ping()
	if err != nil {
		fmt.Println(err)
	}

}
