package main

import (
	"database/sql"
	"fmt"
	//匿名导入默认执行这个包的intit()方法
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	dsn := "root:bjy12345@tcp(127.0.0.1:3306)/sql_test?charset=utf8mb4"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close() //写在err判断下面
	//尝试与数据库建立连接(校验dsn是否真确)
	err = db.Ping()
	if err != nil {
		fmt.Println("connect err:", err)
	}
	println("connect success")

}
