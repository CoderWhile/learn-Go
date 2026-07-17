package model

import (
	"ChatSmall/chatroom/cmd/common/message"
	"net"
)

// 客户端要使用curUser,做成全局的
type CurUser struct {
	Conn net.Conn
	message.User
}
