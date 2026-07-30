package dao

import (
	"GoWeb/bookstore0612/model"
	"fmt"
	"testing"
)

func TestUser(t *testing.T) {
	//fmt.Println("测试User中的函数")
	//t.Run("验证用户名或密码", testLogin)
	//t.Run("验证用户名", testRegister)
	//t.Run("保存用户", testSaveUser)

}

func TestMain(m *testing.M) {
	fmt.Println("测试bookz中的方法")
	m.Run()
}

func TestBook(t *testing.T) {
	fmt.Println("测试bokdao中的相关函数")
	t.Run("测试添加图书", testAddBooks)
}

func testAddBooks(t *testing.T) {
	book := &model.Book{
		Title:     "三国演义",
		Author:    "罗贯中",
		Price:     100,
		Sales:     100,
		Stock:     100,
		ImagePath: "/static/img/default.jpg",
	}
	AddBook(book)

}

func testGetBooks(t *testing.T) {
	books, _ := GetBooks()
	for v, book := range books {
		fmt.Printf("第%d本书是:%v\n", v+1, book)
	}
}
func testLogin(t *testing.T) {

	user, _ := CheckUserNamePassword("admin", "123456")
	fmt.Println("获取的用户信息是:", user)

}
func testRegister(t *testing.T) {

	user, _ := CheckUserName("admin")
	fmt.Println("获取的用户信息是:", user)

}
func testSaveUser(t *testing.T) {
	SaveUser("admin2", "123456", "admin@xupt.com")

}
