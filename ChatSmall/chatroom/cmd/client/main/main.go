package main

import (
	"ChatSmall/chatroom/cmd/client/process"
	"fmt"
	"os"
)

var userId int
var userPwd string
var userName string

func main() {
	var key int
	//var loop = true
	for true {
		fmt.Println("-------------欢迎登陆多人聊天系统-----------")
		fmt.Println("\t\t1.登录")
		fmt.Println("\t\t3.注册")
		fmt.Println("\t\t3.退出")
		fmt.Println("\t\t请选择（1-3）")
		fmt.Scanf("%d\n", &key)
		switch key {
		case 1:
			fmt.Println("登录 ")
			fmt.Println("输入用户名:")
			fmt.Scanf("%s\n", &userName)
			fmt.Println("输入密码：")
			fmt.Scanf("%s\n", &userPwd)
			//1.创建一个UserProcess实例
			up := &process.UserProcess{}
			err := up.Login(userName, userPwd)
			if err != nil {
				fmt.Println(err)
			}
			//loop = false
		case 2:
			fmt.Println("注册")
			fmt.Println("请输入用户昵称：")
			fmt.Scanf("%s\n", &userName)
			fmt.Println("请输入用户密码：")
			fmt.Scanf("%s\n", &userPwd)
			//调用UserProcess。完成注册请求
			up := &process.UserProcess{}
			up.Register(userName, userPwd)
			//loop = false
		case 3:
			fmt.Println("退出")
			os.Exit(0)
			//loop = false
		default:
			fmt.Println("输入有误")
			//loop = false
		}
	}
	//if key == 1 {
	//	//登录
	//	fmt.Println("输入id:")
	//	fmt.Scanf("%d\n", &userId)
	//	fmt.Println("输入密码：")
	//	fmt.Scanf("%s\n", &userPwd)
	//
	//	//
	//
	//	err := login(userId, userPwd)
	//	if err != nil {
	//		fmt.Println("登录失败")
	//	} else {
	//		//fmt.Println("登录成功")
	//	}
	//
	//}
}
