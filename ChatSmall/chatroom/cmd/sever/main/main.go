package main

import (
	"ChatSmall/chatroom/cmd/sever/model"
	process2 "ChatSmall/chatroom/cmd/sever/process"
	"fmt"
	"net"
)

//func readPkg(conn net.Conn) (mes message.Message, err error) {
//	buf := make([]byte, 1024)
//	fmt.Println("读取客户端发送的数据。。")
//	n, err := conn.Read(buf[:4])
//	if n != 4 || err != nil {
//		fmt.Println("conn.Read err:", err)
//		//err = errors.New("read pkg header error")
//		return
//	}
//	//根据读到的buf[:4}转成unit32
//	var pkgLen uint32
//	pkgLen = binary.BigEndian.Uint32(buf[:4])
//	n, err = conn.Read(buf[:pkgLen])
//	if n != int(pkgLen) || err != nil {
//		fmt.Println("conn.Read err:", err)
//		//err = errors.New("read pkg body error")
//		return
//	}
//	//把pkglen反序列化--message.Message
//	err = json.Unmarshal(buf[:pkgLen], &mes)
//	if err != nil {
//		fmt.Println("json.Unmarshal err:", err)
//		return
//	}
//	return
//}
//
//func writePkg(conn net.Conn, data []byte) (err error) {
//	//1.先发送一个长度给对方
//	var pkgLen uint32
//	pkgLen = uint32(len(data))
//	var buf [4]byte
//	binary.BigEndian.PutUint32(buf[0:4], pkgLen)
//	//发送长度
//	var n int
//	n, err = conn.Write(buf[:4])
//	if n != 4 || err != nil {
//		fmt.Println("conn.Write(bytes) err:", err)
//		return
//	}
//	//2.发送消息本身
//	n, err = conn.Write(data)
//	if n != int(pkgLen) || err != nil {
//		fmt.Println("conn.Write(bytes) err:", err)
//		return
//	}
//	return
//}

// 写severProcessLogin专门处理登录请求
//func severProcessLogin(conn net.Conn, mes *message.Message) (err error) {
//	//1.从mes中取出mes.Data,直接反序列化成LoginMes
//	var loginMes message.LoginMes
//	err = json.Unmarshal([]byte(mes.Data), &loginMes)
//	if err != nil {
//		fmt.Println("json.Unmarshal err:", err)
//		return
//	}
//	//先声明一个resMes
//	var resMes message.Message
//	resMes.Type = message.LoginResMesType
//	//再声明一个LoginResMes
//	var loginResMes message.LoginResMes
//	//如果用户基本合法就合法
//	if loginMes.UserId == 100 && loginMes.UserPwd == "123456" {
//		//合法
//		loginResMes.Code = 200
//	} else {
//		//不合法
//		loginResMes.Code = 500 //表示用户不存在
//		loginResMes.Error = "该用户不存在，请注册后再使用"
//	}
//
//	//3.将loginResMes序列化
//	data, err := json.Marshal(loginResMes)
//	if err != nil {
//		fmt.Println("json.Marshal err:", err)
//		return
//	}
//	//4.
//	resMes.Data = string(data)
//	//5.对resMes序列化，准备发送
//	data, err = json.Marshal(resMes)
//	if err != nil {
//		fmt.Println("json.Marshal err:", err)
//		return
//	}
//	//6.发送data，设计到丢包问题，封装到writePkg函数中
//	err = writePkg(conn, data)
//	return
//}

// 写一个SeverProcessMes
// 根据客户端发送消息种类不同，决定调用哪个函数来处理
//func severProcessMes(conn net.Conn, mes *message.Message) (err error) {
//	switch mes.Type {
//	case message.LoginMesType:
//
//		err = severProcessLogin(conn, mes)
//	case message.RegisterMesType:
//	default:
//		fmt.Println("消息类型不存在，无法处理")
//	}
//	return
//}

// 处理和客户端的通讯
func process(conn net.Conn) {
	defer conn.Close()

	//循环读取客户端发送的信息
	//调用总控
	processor := &Processor{
		Conn: conn,
	}
	err := processor.process2()
	if err != nil {
		//
		fmt.Println("客户端和服务器器通讯协程错误=err", err)
	}
	//不管是否正常退出，只要用户已登录就广播下线通知
	if processor.UserId != 0 {
		process2.NotifyUserOffline(processor.UserId)
	}

}

// 完成对userDao的初始化任务
func initUerDao() {
	//初始化顺序问题

	model.MyUserDao = model.NewUserDao(db)
}

func main() {
	//当服务器启动是就启动连接池
	initDB(10, 5)
	initUerDao()
	fmt.Println("服务器在8889端口监听。。")
	listen, err := net.Listen("tcp", ":8889")
	defer listen.Close()
	if err != nil {
		fmt.Println("net.Listen er:", err)
		return
	}
	//监听成功，等待客户端连接服务器
	for {
		fmt.Println("等待客户端连接服务器...")
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listen.Accept error:", err)
			return
		}
		//一旦连接成功，就启动一个协程和客户端保持通讯。。
		go process(conn)
	}
	defer model.MyUserDao.Db.Close()
}
