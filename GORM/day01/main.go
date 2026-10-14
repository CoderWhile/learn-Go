package main

import (
	"database/sql"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UserInfo struct {
	ID     uint
	Name   string
	Gender string
	Hobby  string
}

// 定义模型
type User struct {
	gorm.Model   //内嵌
	Name         string
	Age          sql.NullInt64 //零值类型
	Birthday     *time.Time
	Email        string  `gorm:"type:varchar(100);unique_index"`
	Role         string  `gorm:"size:255"`        //设置字段大小为255
	MemberNumber *string `gorm:"unique;not null"` //设置会员号(member number)
	Num          int     `gorm:"AUTO_INCREMENT"`  //自增类型
	Address      string  `gorm:"index:addr"`      //给该字段创建名为addr的索引
	IgnoreMe     int     `gorm:"-"`               //忽略本字段
}

func main() {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
		return
	}
	//自动迁移
	/*
		如果不存在：自动创建 users 表，
		并根据结构体的字段、类型、Tag（如唯一索引、大小、默认值等）生成建表语句。
		如果已存在：检查表结构与 User 结构体是否匹配。
		如果有差异（如新增了字段、增加了索引），
		它会尝试自动执行 ALTER TABLE 语句来修改表结构
	*/
	db.AutoMigrate(&User{})

	//重命名表
	db.Migrator().RenameTable("users", "newusers2")
	//	db.Table("xiaowangzi").Create(&User{})
}
