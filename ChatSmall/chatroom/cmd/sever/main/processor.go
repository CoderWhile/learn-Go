package main

import (
	"ChatSmall/chatroom/cmd/common/message"
	process2 "ChatSmall/chatroom/cmd/sever/process"
	"ChatSmall/chatroom/cmd/sever/utils"
	"fmt"
	"io"
	"net"
)

// Processor结构体
type Processor struct {
	Conn   net.Conn
	UserId int //记录当前连接的用户ID，用于下线清理
}

func (this *Processor) severProcessMes(mes *message.Message) (err error) {
	fmt.Println("mes:", mes)
	switch mes.Type {
	case message.LoginMesType:
		//创建一个实例UserProcess
		fmt.Println("进入SeverProcessLogin")
		up := &process2.UserProcess{
			Conn: this.Conn,
		}
		err = up.SeverProcessLogin(mes)
		if err != nil {
			fmt.Println("1")
			return err
		}
		//记录登录用户ID，用于下线清理
		this.UserId = up.UserId
	case message.RegisterMesType:
		//创建一个实例UserProcess
		fmt.Println("进入SeverProcessLogin")
		up := &process2.UserProcess{
			Conn: this.Conn,
		}
		err = up.SeverProcessRegister(mes)
	case message.SmsMesType:
		//创建一个SmsProcess实例转发消息
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendGroupMes(mes)
	case message.P2PChatMesType:
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendP2PChat(mes, this.Conn)
	case message.ModifyPwdMesType: //接收到一个用户密码修改的消息
		up := &process2.UserProcess{
			Conn: this.Conn,
		}
		up.ModifyPwdMes(mes)
	case message.UpdateUserNameMesType: //修改用户名
		up := &process2.UserProcess{
			Conn: this.Conn,
		}
		up.UpdateUserName(mes)
	case message.FileSendMesType:
		fmt.Println(1)
		smsProcess := &process2.SmsProcess{}
		smsProcess.SendFileToUser(mes, this.Conn)
	default:
		fmt.Println("消息类型不存在，无法处理")
	}
	return
}
func (this *Processor) process2() (err error) {
	//循环读取数据
	for {
		//将读取数据包封装成一个函数readPkg(),返回Message,Err
		tf := &utils.Transfer{
			Conn: this.Conn,
		}
		var mes message.Message
		mes, err = tf.ReadPkg()
		if err != nil {
			if err == io.EOF {
				fmt.Println("客户端退出，服务器也推出")
				return err
			} else {
				fmt.Println("readPkg err:", err)
				return err
			}
		}
		fmt.Println("mes", mes)
		err = this.severProcessMes(&mes)
		if err != nil {
			return
		}
		//return
	}
	return
}
