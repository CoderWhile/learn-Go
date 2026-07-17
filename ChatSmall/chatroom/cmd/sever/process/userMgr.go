package process2

import (
	"ChatSmall/chatroom/cmd/common/message"
	"ChatSmall/chatroom/cmd/sever/utils"
	"encoding/json"
	"fmt"
)

// UserMgr实例只有一个定义为全局变量
var (
	userMgr *UserMgr
)

type UserMgr struct {
	onlineUsers map[int]*UserProcess
}

// 初始化
func init() {
	userMgr = &UserMgr{
		onlineUsers: make(map[int]*UserProcess, 1024),
	}
}

// 添加完成对onlineUser添加
func (this *UserMgr) AddOnlineuser(up *UserProcess) {
	this.onlineUsers[up.UserId] = up
}

// 删除
func (this *UserMgr) DeleteOnlineUser(userId int) {
	delete(this.onlineUsers, userId)
}

// 查询
func (this *UserMgr) GetAllOnlineUser() map[int]*UserProcess {
	return this.onlineUsers
}

// 根据id返回对应的值
func (this *UserMgr) GetOnlineUserById(userId int) (up *UserProcess, err error) {
	up, ok := this.onlineUsers[userId]
	if !ok {
		//不在线格式化返回
		err = fmt.Errorf("user:%d not exist", userId)
		return
	}
	return
}

// 通知其他用户该用户下线，并从在线列表中删除
func NotifyUserOffline(userId int) {
	//获取下线用户的信息
	up, ok := userMgr.onlineUsers[userId]
	if !ok {
		return
	}
	userName := up.UserName

	fmt.Printf("用户 %d(%s) 已下线\n", userId, userName)

	//组装下线通知消息
	var mes message.Message
	mes.Type = message.NotifyUserStatusMesType

	var notifyUserStatusMes message.NotifyUserStatusMes
	notifyUserStatusMes.UserId = userId
	notifyUserStatusMes.UserName = userName
	notifyUserStatusMes.Status = message.UserOffline

	data, _ := json.Marshal(notifyUserStatusMes)
	mes.Data = string(data)
	data, _ = json.Marshal(mes)

	//通知所有其他在线用户
	for id, up := range userMgr.onlineUsers {
		if id == userId {
			continue
		}
		tf := &utils.Transfer{
			Conn: up.Conn,
		}
		tf.WritePkg(data)
	}

	//从在线列表中删除该用户
	delete(userMgr.onlineUsers, userId)
}
