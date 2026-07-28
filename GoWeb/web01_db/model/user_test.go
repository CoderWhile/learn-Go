package model

import (
	"fmt"
	"testing"
)

// TestMain函数可以再测试函数执行之前做一些其他操作
func TestMain(m *testing.M) {
	fmt.Println("测试开始")
	//通过m.Run()来执行测试函数
	m.Run()
}

func TestUser(t *testing.T) {
	fmt.Println("开始测试Usr中的相关方法")
	//t.Run("测试添加用户:", testAddUser)
	//t.Run("测试获取用户:", testGetUserByID)
	t.Run("测试获取所有用户:", testGetUsers)
}

// 如果函数名不是以Test开头,那么该函数默认不执行,可以将他设置成为一个子测试函数
func testAddUser(t *testing.T) {
	fmt.Println("子测试添加用户:")
	//user := &User{}
	//user.AddUser()
	//user.AddUser2()

}

// 测试获取一个User
func testGetUserByID(t *testing.T) {
	fmt.Println("查询一条记录:")
	user := User{
		ID: 1,
	}
	u, _ := user.GetUserByID()
	fmt.Println("得到的User的信息是:", u)
}

// 测试获取所有user
func testGetUsers(t *testing.T) {
	fmt.Println("测试查询所有记录:")
	user := &User{}
	u, _ := user.GetUsers()
	for k, v := range u {
		fmt.Printf("第%d个用户是%v", k+1, v)
		fmt.Println()
	}
}
