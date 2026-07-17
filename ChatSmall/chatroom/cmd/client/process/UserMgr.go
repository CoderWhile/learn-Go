package process

import (
	"ChatSmall/chatroom/cmd/client/model"
	"ChatSmall/chatroom/cmd/common/message"
	"fmt"
)

// 客户端要维护的Map
var onlineUsers map[int]*message.User = make(map[int]*message.User)
var CurUser model.CurUser //登录成功后完成对初始化
// 初始化工作在登录的时候做

// 客户端显示当前在线的用户
func outputOnlineUser() {
	for id, user := range onlineUsers {
		fmt.Println("id:", id, "username:", user.UserName)
	}
}

// 处理返回的NotifyUserStatusMes
func updateUserstatus(notifyUserStatusMes *message.NotifyUserStatusMes) {

	if notifyUserStatusMes.Status == message.UserOnline {
		//新用户上线，加入在线列表
		user := &message.User{
			UserId:     notifyUserStatusMes.UserId,
			UserName:   notifyUserStatusMes.UserName,
			UserStatus: message.UserOnline,
		}
		onlineUsers[notifyUserStatusMes.UserId] = user
		fmt.Printf("用户 %s(%d) 上线了\n", notifyUserStatusMes.UserName, notifyUserStatusMes.UserId)
	} else if notifyUserStatusMes.Status == message.UserOffline {
		//用户下线，从在线列表中删除
		if _, ok := onlineUsers[notifyUserStatusMes.UserId]; ok {
			fmt.Printf("用户 %s(%d) 下线了\n", notifyUserStatusMes.UserName, notifyUserStatusMes.UserId)
			delete(onlineUsers, notifyUserStatusMes.UserId)
		}
	}

}
