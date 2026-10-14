package main

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 定义模型
type People struct {
	ID   int64
	Name string
	Age  int64
}

type User struct {
	gorm.Model
	Name string
	Age  int64
}

func main() {
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), //打印每条sql
	})
	if err != nil {
		panic(err)
		return
	}
	//2.把模型与数据库中的模型对应起来
	//if err := db.AutoMigrate(&People{}); err != nil {
	//
	//	log.Fatal(err)
	//}
	//
	//db.AutoMigrate(&User{})
	////创建
	//u1 := User{Name: "Wang", Age: 18}
	//db.Debug().Create(&u1)
	//u2 := User{Name: "Xumin", Age: 19}
	//db.Debug().Create(&u2)
	//3.创建
	//u := People{Name: "bai", Age: 19} //在代码层面创建一个Use对象
	////判断主键是否为空
	////	result := db.Create(&u)
	//db.Debug().Create(&u)
	////fmt.Println("err:", result.Error)                 //错误
	////fmt.Println("rowsAffected:", result.RowsAffected) //影响行数
	//fmt.Println("u.ID:", u.ID) //主键被回写

	//查询:
	var u3 User
	db.First(&u3)
	fmt.Printf("user:%#v\n", u3)

	//更新:
	
}
