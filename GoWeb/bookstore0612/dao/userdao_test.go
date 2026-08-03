package dao

import (
	"GoWeb/bookstore0612/model"
	"fmt"
	"testing"
)

//func TestUser(t *testing.T) {
//	//fmt.Println("测试User中的函数")
//	//t.Run("验证用户名或密码", testLogin)
//	//t.Run("验证用户名", testRegister)
//	//t.Run("保存用户", testSaveUser)
//
//}

func TestMain(m *testing.M) {
	fmt.Println("测试bookz中的方法")
	m.Run()
}

//
//func TestBook(t *testing.T) {
//	fmt.Println("测试bokdao中的相关函数")
//	t.Run("测试添加图书", testGetBooks)
//}

//	func testAddBooks(t *testing.T) {
//		book := &model.Book{
//			Title:     "三国演义",
//			Author:    "罗贯中",
//			Price:     100,
//			Sales:     100,
//			Stock:     100,
//			ImagePath: "/static/img/default.jpg",
//		}
//		AddBook(book)
//
// }
//
//	func testGetBooks(t *testing.T) {
//		books, _ := GetPageBooksByPrice("1", "0", "100")
//		fmt.Printf("books: %+v\n", books)
//		for v, book := range books.Books {
//			fmt.Printf("第%d本书是:%v\n", v+1, book)
//		}
//	}
//
// func testLogin(t *testing.T) {
//
//	user, _ := CheckUserNamePassword("admin", "123456")
//	fmt.Println("获取的用户信息是:", user)
//
// }
// func testRegister(t *testing.T) {
//
//	user, _ := CheckUserName("admin")
//	fmt.Println("获取的用户信息是:", user)
//
// }
//
//	func testSaveUser(t *testing.T) {
//		SaveUser("admin2", "123456", "admin@xupt.com")
//
// }
func TestSession(t *testing.T) {
	fmt.Println("测试Session相关函数")
	//	t.Run("测试添加Session", testAddSession)
	t.Run("测试删除Session", testDeleteSession)
}
func testAddSession(t *testing.T) {
	sess := &model.Session{
		SessionID: "16545435154",
		UserName:  "王佳乐",
		UserID:    15,
	}
	AddSession(sess)
	DeleteSession("15")
}
func testDeleteSession(t *testing.T) {
	DeleteSession("16545435154")
}
