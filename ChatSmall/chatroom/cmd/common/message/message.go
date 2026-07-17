package message

// 确定一些消息类型
const (
	LoginMesType            = "LoginMes"
	LoginResMesType         = "LoginResMes"
	RegisterMesType         = "RegisterMes"
	RegisterResMesType      = "RegisterResMes"
	NotifyUserStatusMesType = "NotifyUserStatusMes"
	SmsMesType              = "SmsMes"
	P2PChatMesType          = "P2PChatMes"
	P2PChatResMesType       = "P2PChatResMes"
	ModifyPwdMesType        = "ModifyPwdMes"
	ModidyPwdResMesType
	UpdateUserNameMesType    = "UpdateUserNameMes"
	UpdateUserNameResMesType = "UpdateUserNameResMes"
	FileSendMesType          = "FileSendMes"
	FileSendResMesType       = "FileSendResMes"
)

type Message struct {
	Type string `json:"type"` //消息类型
	Data string `json:"data"` //消息的内容
}

// 用户状态常量
const (
	UserOnline = iota
	UserOffline
	UserBusyStatus
)

// 定义两个消息
type LoginMes struct {
	UserId   int    `json:"userId"`
	UserPwd  string `json:"userPwd"`
	UserName string `json:"userName"`
}

type LoginResMes struct {
	UserId    int      `json:"userId"`    //成功登录的id
	Code      int      `json:"code"`      //返回状态码，500表示用户未注册，200表示登录成功
	UserIds   []int    `json:"userIds"`   //增加字段，保存用户id切片
	UserNames []string `json:"userNames"` //增加用户名字字段返回用户的名字
	Error     string   `json:"error"`     //返回错误信息
}
type RegisterMes struct {
	User User `json:"user"` //类型就是User结构体
}
type RegisterResMes struct {
	Code  int    `json:"code"` //400:用户已经占用，200注册成功
	Error string `json:"error"`
}

// 服务器端推送用户状态变化消息
type NotifyUserStatusMes struct {
	UserId   int    `json:"userId"` //
	UserName string `json:"userName"`
	Status   int    `json:"Status"` //用户状态
}

// 增加一个SmsMes发送的
type SmsMes struct {
	Content string `json:"content"`
	User           //匿名结构体
}

// 点对点私聊消息
type P2PChatMes struct {
	Content  string `json:"content"`
	TargetId int    `json:"targetId"` //目标用户id
	User            //发送者信息
}

// 私聊响应消息
type P2PChatResMes struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

//用户密码修改请求消息

type ModifyPwdMes struct {
	NewPassword string `json:"newPassword"`
	OldPassword string `json:"oldPassword"`
	User               //发送者信息
}
type ModidyPwdResMes struct {
	Code  int    `json:"code"`
	Error string `json:"error"`
}

// 修改用户名请求消息
type UpdateUserNameMes struct {
	NewUserName string `json:"newUserName"`
	Password    string `json:"password"` //验证身份
	User               //发送者信息
}

// 修改用户名响应消息
type UpdateUserNameResMes struct {
	Code    int    `json:"code"`
	NewName string `json:"newName"` //修改后的新用户名
	Error   string `json:"error"`
}

// 文件分块消息
type FileSendMes struct {
	TargetId   int    `json:"targetId"`
	FileName   string `json:"fileName"`
	FileSize   int    `json:"fileSize"`
	TotalChunk int    `json:"totalChunk"`
	ChunkIndex int    `json:"chunkIndex"` //当前是第几快
	Data       string `json:"data"`
	User              //发送者的信息
}

type FileSendResMes struct {
	Code     int    `json:"code"`
	FileName string `json:"fileName"`
	Error    string `json:"error"`
}
