package process

import (
	"ChatSmall/chatroom/cmd/client/utils"
	"ChatSmall/chatroom/cmd/common/message"
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
)

// 显示登录成功后的界面
func ShowMenu(username string) {
	fmt.Printf("---------%s用户登录界面----------\n", username)
	fmt.Println("---------1.显示在线用户列表--------")
	fmt.Println("---------2.发送消息---------------")
	fmt.Println("---------3.信息列表---------------")
	fmt.Println("---------4.发送文件---------------")
	fmt.Println("---------5.私聊------------------")
	fmt.Println("---------6.修改密码---------------")
	fmt.Println("---------7.修改用户名---------------")
	fmt.Println("---------8.退出系统---------------")
	var key int
	fmt.Scanf("%d\n", &key)
	var content string
	smsProcess := &SmsProcess{}
	userProcess := &UserInfoMgr{}
	switch key {
	case 1:
		fmt.Println("显示在线用户列表")
		outputOnlineUser()
	case 2:
		fmt.Println("输入要发送的消息")
		fmt.Scanf("%s\n", &content)
		smsProcess.SendGroupMes(content)
	case 3:
		fmt.Println("信息列表")
	case 4:
		fmt.Println("请输入要发送目标用户id")
		var targetid int
		fmt.Scanf("%d\n", &targetid)
		fmt.Println("请输入要发送的文件地址")
		var filePath string
		fmt.Scanf("%s\n", &filePath)
		smsProcess.SendFile(targetid, filePath)

	case 5:
		fmt.Println("请输入目标用户ID：")
		var targetId int
		fmt.Scanf("%d\n", &targetId)
		fmt.Println("请输入私聊内容：")
		var p2pContent string
		fmt.Scanf("%s\n", &p2pContent)
		smsProcess.SendP2PChat(p2pContent, targetId)
	case 6: //修改密码
		//输入原密码，输入两次新密码
		//输入原密码进行判断，、
		var oldPassword string
		var newPassword1 string
		var newPassword2 string
		fmt.Println("请输入原用户密码：")
		fmt.Scanf("%s\n", &oldPassword)
		fmt.Println("请输入新密码：")
		fmt.Scanf("%s\n", &newPassword1)
		fmt.Println("请再次输入新密码")
		fmt.Scanf("%s\n", &newPassword2)
		//输入新密码
		if newPassword2 != newPassword1 {
			fmt.Println("两次新密码输入不一样")

		} else {
			userProcess.ModifyUserPwd(oldPassword, newPassword1)
		}
	case 7: //修改用户名
		var newUserName string
		var userNamePwd string
		fmt.Println("请输入新用户名：")
		fmt.Scanf("%s\n", &newUserName)
		fmt.Println("请输入密码验证身份：")
		fmt.Scanf("%s\n", &userNamePwd)
		userProcess.UpdateUserName(newUserName, userNamePwd)
	case 8:
		fmt.Println("退出系统")
		os.Exit(0)
	default:
		fmt.Println("输入错误。")

	}
}

// 和服务器端保持通讯
func severProcessMes(conn net.Conn) {
	//创建transfer实例，不停的读取服务器发送的消息
	tf := &utils.Transfer{
		Conn: conn,
	}
	for {
		mes, err := tf.ReadPkg()
		if err != nil {
			fmt.Println("11readPkg err:", err)
			return
		}
		//读取到消息，下一步处理
		//fmt.Println("mes", mes)
		switch mes.Type {
		case message.NotifyUserStatusMesType: //有人上线了
			//处理
			//取出NotifyUserStatusMes
			var notifyUserStatusMes message.NotifyUserStatusMes
			json.Unmarshal([]byte(mes.Data), &notifyUserStatusMes)
			updateUserstatus(&notifyUserStatusMes)

		//把这个人的用户的信息加入到客户端维护的map【int]User
		case message.SmsMesType: //有人群发消息了
			outputGroupMes(&mes)
		case message.P2PChatMesType: //有人私聊你了
			var p2pMes message.P2PChatMes
			json.Unmarshal([]byte(mes.Data), &p2pMes)
			fmt.Printf("\n用户 %d(%s) 对你私聊：%s\n", p2pMes.UserId, p2pMes.UserName, p2pMes.Content)
		case message.P2PChatResMesType: //私聊发送结果
			var p2pResMes message.P2PChatResMes
			json.Unmarshal([]byte(mes.Data), &p2pResMes)
			if p2pResMes.Code == 200 {
				fmt.Println("私聊发送成功")
			} else {
				fmt.Println("私聊发送失败：", p2pResMes.Error)
			}
		case message.ModidyPwdResMesType: //接收到密码修改是否完成的消息
			var modifyPwdResMes message.ModidyPwdResMes
			json.Unmarshal([]byte(mes.Data), &modifyPwdResMes)
			if modifyPwdResMes.Code == 200 {
				fmt.Println("密码修改成功")
			} else {
				fmt.Println("密码修改失败err:", modifyPwdResMes.Error)
			}
		case message.UpdateUserNameResMesType: //接收到用户名修改完成的消息
			var updateResMes message.UpdateUserNameResMes
			json.Unmarshal([]byte(mes.Data), &updateResMes)
			if updateResMes.Code == 200 {
				CurUser.UserName = updateResMes.NewName
				fmt.Println("用户名修改成功，新用户名：", updateResMes.NewName)
			} else {
				fmt.Println("用户名修改失败：", updateResMes.Error)
			}
		case message.FileSendMesType:
			//接收到一个用户的文件
			var fileSendMes message.FileSendMes
			json.Unmarshal([]byte(mes.Data), &fileSendMes)
			newFilepath := fmt.Sprintf("[%s]%s", CurUser.UserName, filepath.Base(fileSendMes.FileName))
			var file *os.File
			if fileSendMes.ChunkIndex == 0 {
				file, err = os.Create(newFilepath)
				if err != nil {
					fmt.Println(err)
				}
			} else {
				file, err = os.OpenFile(newFilepath, os.O_APPEND|os.O_WRONLY, 0777)
				if err != nil {
					fmt.Println(err)
				}
			}

			w := bufio.NewWriter(file)

			decoded, err := base64.StdEncoding.DecodeString(fileSendMes.Data)
			if err != nil {
				log.Fatal("invalid base64:", err)
			}
			w.Write(decoded)
			w.Flush()
			file.Close()
			if fileSendMes.ChunkIndex == fileSendMes.TotalChunk-1 {
				fmt.Printf("文件接收完毕来自%s\n", fileSendMes.FileName)
				w.Flush()
				file.Close()
			}
		case message.FileSendResMesType:
			var fileSendResMes message.FileSendResMes
			json.Unmarshal([]byte(mes.Data), &fileSendResMes)
			if fileSendResMes.Code == 200 {
				fmt.Println("文件发送成功")
			} else {
				fmt.Println("文件发送失败")
			}

		default:
			fmt.Println("服务器端返回了一个未知的消息类型")
		}

	}
}
