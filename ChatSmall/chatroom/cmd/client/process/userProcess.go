package process

import (
	"ChatSmall/chatroom/cmd/client/utils"
	"ChatSmall/chatroom/cmd/common/message"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type UserProcess struct {
	//字段。。。

}

func (this *UserProcess) Login(userName string, userPwd string) (err error) {
	//1.连接服务器
	conn, err := net.Dial("tcp", "localhost:8889")
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return
	}
	defer conn.Close()
	//准备conn发送消息给服务
	var mes message.Message
	mes.Type = message.LoginMesType
	//3.创建一个LoginMes结构体
	var loginMes message.LoginMes
	//loginMes.UserId = userId
	loginMes.UserName = userName
	loginMes.UserPwd = userPwd
	//4.讲loginMes序列化
	data, err := json.Marshal(loginMes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//5.吧data赋给，mex.Data字段
	mes.Data = string(data)
	//6.将mex进行序列化
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//7data就是要发送的消息
	//7.1先发送data的长度给服务器
	//先获取data的长度-》转为切片
	var pkgLen uint32
	pkgLen = uint32(len(data))
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[0:4], pkgLen)
	//发送长度
	var n, err2 = conn.Write(buf[:4])
	if n != 4 || err2 != nil {
		fmt.Println("conn.Write(bytes) err:", err)
		return
	}
	//fmt.Printf("客户端发送消息的长度成功长度为：%d,内容：%s\n", len(data), string(data))
	//发送消息本身
	_, err2 = conn.Write(data)
	if err2 != nil {
		fmt.Println("conn.Write(bytes) err:", err)
		return
	}
	//这里还需要处理服务器端返回的消息
	tf := &utils.Transfer{
		Conn: conn,
	}
	mes, err = tf.ReadPkg() //me
	if err != nil {
		fmt.Println("1readPkg err:", err)
		return
	}
	//将mes的data部分反序列化成LoginResMes
	var loginResMes message.LoginResMes
	err = json.Unmarshal([]byte(mes.Data), &loginResMes)
	if loginResMes.Code == 200 {
		//初始huaCurUser
		CurUser.Conn = conn
		CurUser.UserId = loginResMes.UserId
		CurUser.UserName = userName
		CurUser.UserStatus = message.UserOnline
		fmt.Println("登录成功")
		//显示当前在线用户列表
		//通过登录后服务器在登录信息返回包中获取当前在线用户的id
		//	fmt.Println("当前在线用户列表如下：")
		for i, v := range loginResMes.UserIds {

			//fmt.Printf("用户id:%d\n", v)
			//完成客户端onlineUser的初始化
			user := &message.User{
				UserId:     v,
				UserStatus: message.UserOnline,
			}
			onlineUsers[v] = user
			onlineUsers[v].UserName = loginResMes.UserNames[i]
		}

		fmt.Println()
		fmt.Println()

		//在客户端起一个协程
		//保持和服务器端的通讯，服务器有数据推送给客户端
		//可以接收并显示客户端的终端

		go severProcessMes(conn)
		//调用登录成功后的菜单（循环）
		for {
			ShowMenu(CurUser.UserName)
		}
	} else {
		fmt.Println(loginResMes.Error)
	}
	return
}

func (this *UserProcess) Register(userName, userPwd string) (err error) {
	//1.连接服务器
	conn, err := net.Dial("tcp", "localhost:8889")
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return
	}
	defer conn.Close()
	//准备conn发送消息给服务
	var mes message.Message
	mes.Type = message.RegisterMesType
	//3.创建一个LoginMes结构体
	var RegisterMes message.RegisterMes
	//RegisterMes.User.UserId = userId
	RegisterMes.User.UserPwd = userPwd
	RegisterMes.User.UserName = userName
	//4将Register序列化
	data, err := json.Marshal(RegisterMes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//5
	//5.吧data赋给，mex.Data字段
	mes.Data = string(data)
	//6.将mex进行序列化
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//7.
	tf := &utils.Transfer{
		Conn: conn,
	}
	//发送data给服务器
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("注册信息发送错误err:", err)
	}
	mes, err = tf.ReadPkg() //j就是ReginterResMes
	if err != nil {
		fmt.Println("1readPkg err:", err)
		return
	}

	//将mes的data部分反序列化成ReginterResMes
	var registerResMes message.RegisterResMes
	err = json.Unmarshal([]byte(mes.Data), &registerResMes)
	if registerResMes.Code == 200 {
		fmt.Println("注册成功，重新登录")
		os.Exit(0)
	} else {
		fmt.Println(registerResMes.Error)
		os.Exit(0)
	}

	return
}
