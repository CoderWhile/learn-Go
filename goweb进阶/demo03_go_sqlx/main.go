package main

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

type User struct {
	ID   int    `db:"id"`
	Age  int    `db:"age"`
	Name string `db:"name"`
}

func initDB() (err error) {
	dsn := "root:bjy12345@tcp(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True"
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		fmt.Printf("connect DB failed:%v\n", err)
		return
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	return
}

func insertUserDemo() (err error) {
	sqlStr := "INSERT INTO users (name,age) VALUES (:name,:age)"
	_, err = db.NamedExec(sqlStr, map[string]interface{}{
		"name": "鸭鸭",
		"age":  "21",
	})
	return
}
func selectUserDemo() (err error) {
	sqlStr := "select id,age,name from users"
	var users []User
	err = db.Select(&users, sqlStr)
	if err != nil {
		fmt.Printf("query failed:%v\n", err)
		return
	}
	fmt.Printf("users:%v\n", users)
	return
}

func transactionDemo2() (err error) {
	tx, err := db.Beginx() // 开启事务
	if err != nil {
		fmt.Printf("begin trans failed, err:%v\n", err)
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			fmt.Println("rollback")
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
			fmt.Println("commit")
		}
	}()

	sqlStr1 := "Update user set age=20 where id=?"

	rs, err := tx.Exec(sqlStr1, 1)
	if err != nil {
		return err
	}
	n, err := rs.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed")
	}
	sqlStr2 := "Update user set age=50 where i=3"
	rs, err = tx.Exec(sqlStr2, 5)
	if err != nil {
		return err
	}
	n, err = rs.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed")
	}
	return err
}

// BatchInsertUsers 自行构造批量插入的语句
func BatchInsertUsers(users []*User) error {
	// 存放 (?, ?) 的slice
	valueStrings := make([]string, 0, len(users))
	// 存放values的slice
	valueArgs := make([]interface{}, 0, len(users)*2)
	// 遍历users准备相关数据
	for _, u := range users {
		// 此处占位符要与插入值的个数对应
		valueStrings = append(valueStrings, "(?, ?)")
		valueArgs = append(valueArgs, u.Name)
		valueArgs = append(valueArgs, u.Age)
	}
	// 自行拼接要执行的具体语句
	stmt := fmt.Sprintf("INSERT INTO user (name, age) VALUES %s",
		strings.Join(valueStrings, ","))
	_, err := db.Exec(stmt, valueArgs...)
	return err
}

// 想使用sqlx.In实现批量插入
// 前提实现driver.Valuer接口
func (u User) Value() (driver.Value, error) {
	return []interface{}{u.Name, u.Age}, nil

}
// BatchInsertUsers2 使用sqlx.In + driver.Valuer实现批量插入
func BatchInsertUsers2(users []*User) error {
	// 将[]*User转换为[]interface{}
	valueArgs := make([]interface{}, 0, len(users))
	for _, u := range users {
		valueArgs = append(valueArgs, *u) // 传值类型，因为Value()接收者是值类型
	}

	// 构建占位符，每个User对象对应一个(?)
	placeholders := make([]string, 0, len(users))
	for range users {
		placeholders = append(placeholders, "(?)")
	}

	// sqlx.In会调用User.Value()方法展开每个User对象
	query, args, err := sqlx.In(
		fmt.Sprintf("INSERT INTO users (name, age) VALUES %s", strings.Join(placeholders, ",")),
		valueArgs...,
	)
	if err != nil {
		return err
	}
	fmt.Println("生成的SQL:", query)
	fmt.Println("参数:", args)
	_, err = db.Exec(query, args...)
	return err
}

// BatchInsertUsers3 使用sqlx.In直接传递参数（更简单的方式）
func BatchInsertUsers3(users []*User) error {
	// 手动将每个User的字段展开为参数
	valueArgs := make([]interface{}, 0, len(users)*2)
	for _, u := range users {
		valueArgs = append(valueArgs, u.Name, u.Age)
	}

	// 构建占位符
	placeholders := make([]string, 0, len(users))
	for range users {
		placeholders = append(placeholders, "(?, ?)")
	}

	query := fmt.Sprintf("INSERT INTO users (name, age) VALUES %s", strings.Join(placeholders, ","))
	fmt.Println("生成的SQL:", query)
	fmt.Println("参数:", valueArgs)
	_, err := db.Exec(query, valueArgs...)
	return err
}

func main() {
	if err := initDB(); err != nil {
		fmt.Printf("init DB failed:%v\n", err)
		return
	}

	// 测试批量插入
	users := []*User{
		{Name: "咩咩", Age: 88},
		{Name: "喵喵", Age: 10},
		{Name: "汪汪", Age: 25},
	}

	// 使用方法2：使用sqlx.In + driver.Valuer
	err := BatchInsertUsers2(users)
	if err != nil {
		fmt.Printf("batch insert failed:%v\n", err)
		return
	}

	fmt.Println("batch insert success")

	// 查询验证
	selectUserDemo()
}
