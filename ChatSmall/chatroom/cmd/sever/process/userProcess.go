package process2

import (
	"ChatSmall/chatroom/cmd/common/message"
	"ChatSmall/chatroom/cmd/sever/model"
	"ChatSmall/chatroom/cmd/sever/utils"
	"encoding/json"
	"fmt"
	"net"
)

type UserProcess struct {
	//有什么字段？连接，
	Conn net.Conn
	//表示是哪个用户的
	UserId   int
	UserName string
}

// 通知锁在线用户的方法
// userId要通知其他在线用户我上线了
func (this *UserProcess) NotifyOtherOnlineUser(userId int, userName string) {
	//遍历onlineUser，然后一个个发送NofyUserstatusMe
	for id, up := range userMgr.onlineUsers {
		if id == userId {
			continue
		}
		//开始通知
		//方法
		//up是各个在线用户的UserProcess
		up.NotifyMeOnline(userId, userName)
	}
}
func (this *UserProcess) NotifyMeOnline(userId int, userName string) {
	//组装通知上线消息
	var mes message.Message
	mes.Type = message.NotifyUserStatusMesType

	var notifyUserStatusMes message.NotifyUserStatusMes
	notifyUserStatusMes.UserId = userId
	notifyUserStatusMes.UserName = userName
	notifyUserStatusMes.Status = message.UserOnline

	//对序列化
	data, err := json.Marshal(notifyUserStatusMes)
	if err != nil {
		fmt.Println("notifyUserStatusMes json marshal err:", err)
		return
	} //序列化后的notifyUserStatusMes赋值给mes.Data
	mes.Data = string(data)
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("notifyUserStatusMes2 json marshal err:", err)
		return
	}
	//发送创建transfer实例，发送
	tf := &utils.Transfer{
		Conn: this.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("notifyOnline json write err:", err)
		return
	}
}

// 登录
func (this *UserProcess) SeverProcessLogin(mes *message.Message) (err error) {
	//1.从mes中取出mes.Data,直接反序列化成LoginMes
	var loginMes message.LoginMes
	err = json.Unmarshal([]byte(mes.Data), &loginMes)
	if err != nil {
		fmt.Println("json.Unmarshal err:", err)
		return
	}
	fmt.Println("反序列化loginMes成功")
	//先声明一个resMes
	var resMes message.Message
	resMes.Type = message.LoginResMesType
	//再声明一个LoginResMes
	var loginResMes message.LoginResMes
	//到数据库完成验证
	//使用model.MyUserDao到数据库中验证

	user, err := model.MyUserDao.Login(loginMes.UserName, loginMes.UserPwd)

	//
	if err != nil {
		if err == model.ERROR_USER_NOTEXISTS {
			loginResMes.Code = 500
			loginResMes.Error = err.Error()
		} else if err == model.ERROR_USER_PWD {
			loginResMes.Code = 403
			loginResMes.Error = err.Error()
		} else {
			loginResMes.Code = 505
			loginResMes.Error = "服务器内部错误。。"
		}
		//返回具体错误信息
	} else {
		fmt.Println("每次登录读取到的user:", user)
		loginResMes.Code = 200
		//用户登录成功，维护在线用户
		//把成功登录的用户，放到userMgr中
		//将登录成功的用户的userId赋给this
		this.UserName = user.UserName
		this.UserId = user.UserId
		userMgr.AddOnlineuser(this)
		//通知其他用户我上线了
		this.NotifyOtherOnlineUser(user.UserId, user.UserName)
		//将当前在线用户Id放入loginResMes.UsersId
		loginResMes.UserId = user.UserId
		for id, up := range userMgr.onlineUsers {
			loginResMes.UserIds = append(loginResMes.UserIds, id)
			loginResMes.UserNames = append(loginResMes.UserNames, up.UserName)
		}
		fmt.Println(user, "登录成功")
	}

	//如果用户基本合法就合法
	//if loginMes.UserId == 100 && loginMes.UserPwd == "123456" {
	//	//合法
	//	loginResMes.Code = 200
	//	fmt.Println(loginMes.UserId, "登录成功")
	//} else {
	//	//不合法
	//	loginResMes.Code = 500 //表示用户不存在
	//	loginResMes.Error = "该用户不存在，请注册后再使用"
	//}

	//3.将loginResMes序列化
	data, err := json.Marshal(loginResMes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//4.
	resMes.Data = string(data)
	//5.对resMes序列化，准备发送
	data, err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//6.发送data，设计到丢包问题，封装到writePkg函数中
	//先创建transger
	tf := &utils.Transfer{
		Conn: this.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("客户端返回信息write err:", err)
	}
	return
}

// 注册
func (this *UserProcess) SeverProcessRegister(mes *message.Message) (err error) {
	//取出mes的data序列化为RegisterMes
	var registerMes message.RegisterMes
	err = json.Unmarshal([]byte(mes.Data), &registerMes)
	if err != nil {
		fmt.Println("RegisterMes json.Unmarshal err:", err)
		return
	}
	fmt.Println("反序列化registerMes成功")
	//先声明一个resMes,将相应返回给客户端
	var resMes message.Message
	resMes.Type = message.RegisterResMesType
	//再声明一个registerResMes
	var registerResMes message.RegisterResMes
	//到数据库完成注册
	err = model.MyUserDao.Register(&registerMes.User)
	if err != nil {
		if err == model.ERROR_USER_EXISTS {
			registerResMes.Code = 505
			registerResMes.Error = model.ERROR_USER_EXISTS.Error()
		} else {
			registerResMes.Code = 506
			registerResMes.Error = "注册发生未知错误"
		}
	} else {
		registerResMes.Code = 200
	}

	//3.将registerResMes序列化
	data, err := json.Marshal(registerResMes)
	if err != nil {
		fmt.Println("registerResMes json.Marshal err:", err)
		return
	}
	//4.
	resMes.Data = string(data)
	//5.对resMes序列化，准备发送
	data, err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//6.发送data，设计到丢包问题，封装到writePkg函数中
	//先创建transger
	tf := &utils.Transfer{
		Conn: this.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("客户端返回注册信息write err:", err)
	}
	return
}

// 密码修改
func (this *UserProcess) ModifyPwdMes(mes *message.Message) (err error) {
	var modifyMes message.ModifyPwdMes
	err = json.Unmarshal([]byte(mes.Data), &modifyMes)
	if err != nil {
		fmt.Println("ModifyPwdMes json.Unmarshal err:", err)
	}
	//拿到修改信息
	//到数据库完成是否能修改密码
	err, _ = model.MyUserDao.ModifyPwd(modifyMes.NewPassword, &modifyMes.User)
	if err != nil {
		return err
	}
	//包装密码是否修改成功的消息给客户端

	//先声明一个resMes,将相应返回给客户端
	var resMes message.Message
	resMes.Type = message.ModidyPwdResMesType
	//再声明一个registerResMes
	var modifyPwdResMes message.ModidyPwdResMes

	if err != nil {
		if err == model.ERROR_USER_EXISTS {
			modifyPwdResMes.Code = 505
			modifyPwdResMes.Error = model.ERROR_USER_PWD.Error()
		} else {
			modifyPwdResMes.Code = 506
			modifyPwdResMes.Error = "修改密码发生未知错误"
		}
	} else {
		modifyPwdResMes.Code = 200
	}

	//3.将registerResMes序列化
	data, err := json.Marshal(modifyPwdResMes)
	if err != nil {
		fmt.Println("registerResMes json.Marshal err:", err)
		return
	}
	//4.
	resMes.Data = string(data)
	//5.对resMes序列化，准备发送
	data, err = json.Marshal(resMes)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}
	//6.发送data，设计到丢包问题，封装到writePkg函数中
	//先创建transger
	tf := &utils.Transfer{
		Conn: this.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("客户端返回密码修改信息write err:", err)
	}
	return
}

// 修改用户名
func (this *UserProcess) UpdateUserName(mes *message.Message) (err error) {
	//解析请求
	var updateMes message.UpdateUserNameMes
	err = json.Unmarshal([]byte(mes.Data), &updateMes)
	if err != nil {
		fmt.Println("UpdateUserNameMes json.Unmarshal err:", err)
		return
	}

	//调用DAO修改用户名
	err = model.MyUserDao.UpdateUserName(updateMes.UserId, updateMes.NewUserName, updateMes.Password)

	//包装响应消息
	var resMes message.Message
	resMes.Type = message.UpdateUserNameResMesType

	var updateResMes message.UpdateUserNameResMes

	if err != nil {
		if err == model.ERROR_USER_PWD {
			updateResMes.Code = 403
			updateResMes.Error = "密码不正确"
		} else if err == model.ERROR_USER_EXISTS {
			updateResMes.Code = 505
			updateResMes.Error = "用户名已被占用"
		} else if err == model.ERROR_USER_NOTEXISTS {
			updateResMes.Code = 500
			updateResMes.Error = "用户不存在"
		} else {
			updateResMes.Code = 506
			updateResMes.Error = "修改用户名发生未知错误"
		}
	} else {
		updateResMes.Code = 200
		updateResMes.NewName = updateMes.NewUserName
		//更新服务端UserProcess中的用户名
		this.UserName = updateMes.NewUserName
	}

	data, _ := json.Marshal(updateResMes)
	resMes.Data = string(data)
	data, _ = json.Marshal(resMes)

	tf := &utils.Transfer{
		Conn: this.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("客户端返回修改用户名信息write err:", err)
	}
	return
}
