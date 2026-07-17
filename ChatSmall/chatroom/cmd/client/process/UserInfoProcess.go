package process

import (
	"ChatSmall/chatroom/cmd/client/utils"
	"ChatSmall/chatroom/cmd/common/message"
	"encoding/json"
	"fmt"
)

// 修改用户密码
type UserInfoMgr struct {
}

// 修改密码
func (this *UserInfoMgr) ModifyUserPwd(oldPwd, newPwd string) (err error) {
	//创建一个Message
	var mes message.Message
	mes.Type = message.ModifyPwdMesType

	//P2PChatMes实例
	var modifyPwdMes message.ModifyPwdMes
	modifyPwdMes.NewPassword = newPwd
	modifyPwdMes.OldPassword = oldPwd
	modifyPwdMes.UserId = CurUser.UserId
	modifyPwdMes.UserName = CurUser.UserName
	modifyPwdMes.UserStatus = CurUser.UserStatus

	//序列化p2pMes
	data, err := json.Marshal(modifyPwdMes)
	if err != nil {
		fmt.Println("P2PChatMes json.Marshal err:", err)
		return
	}
	mes.Data = string(data)
	//序列化mes
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("P2PChatMes2 json.Marshal err:", err)
		return
	}
	//发送
	tf := utils.Transfer{
		Conn: CurUser.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("客户端修改密码 WritePkg err:", err)
		return
	}
	return
}

// 修改用户名
func (this *UserInfoMgr) UpdateUserName(newUserName, password string) (err error) {
	//创建一个Message
	var mes message.Message
	mes.Type = message.UpdateUserNameMesType

	//UpdateUserNameMes实例
	var updateMes message.UpdateUserNameMes
	updateMes.NewUserName = newUserName
	updateMes.Password = password
	updateMes.UserId = CurUser.UserId
	updateMes.UserName = CurUser.UserName
	updateMes.UserStatus = CurUser.UserStatus

	//序列化
	data, err := json.Marshal(updateMes)
	if err != nil {
		fmt.Println("UpdateUserNameMes json.Marshal err:", err)
		return
	}
	mes.Data = string(data)
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("UpdateUserNameMes2 json.Marshal err:", err)
		return
	}
	//发送
	tf := utils.Transfer{
		Conn: CurUser.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("客户端修改用户名 WritePkg err:", err)
		return
	}
	return
}
