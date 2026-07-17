package process2

import (
	"ChatSmall/chatroom/cmd/common/message"
	"ChatSmall/chatroom/cmd/sever/utils"
	"encoding/json"
	"fmt"
	"net"
)

type SmsProcess struct {
}

// 转发消息
func (this *SmsProcess) SendGroupMes(mes *message.Message) {
	//遍历服务器端的onlineUser Map
	//将消息转发出去：//
	//取出mes的内容
	var smsMes message.SmsMes
	err := json.Unmarshal([]byte(mes.Data), &smsMes)
	if err != nil {
		fmt.Println("json,Unmarshal err:", err)
		return
	}
	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("json,Marshal err:", err)
	}
	for id, up := range userMgr.onlineUsers {
		if id == smsMes.UserId {
			continue
		}
		this.SendMesToEachOnlineUser(data, up.Conn)
	}
}
func (this *SmsProcess) SendMesToEachOnlineUser(data []byte, conn net.Conn) {
	tf := &utils.Transfer{
		Conn: conn,
	}
	err := tf.WritePkg(data)
	if err != nil {
		fmt.Println("转发消息失败 err:", err)
	}
}

// 发送私聊消息
func (this *SmsProcess) SendP2PChat(mes *message.Message, senderConn net.Conn) {
	//解析私聊消息
	var p2pMes message.P2PChatMes
	err := json.Unmarshal([]byte(mes.Data), &p2pMes)
	if err != nil {
		fmt.Println("SendP2PChat json.Unmarshal err:", err)
		return
	}

	//查找目标用户是否在线
	targetUp, err := userMgr.GetOnlineUserById(p2pMes.TargetId)
	if err != nil {
		fmt.Println("目标用户不在线:", err)
		this.sendP2PRes(senderConn, 404, "目标用户不在线")
		return
	}
	fmt.Printf("用户 %s 向用户 %s 发送私聊消息：%s\n", p2pMes.UserName, targetUp.UserName, p2pMes.Content)
	//转发原消息给目标用户
	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("SendP2PChat json.Marshal err:", err)
		return
	}
	this.SendMesToEachOnlineUser(data, targetUp.Conn)

	//通知发送者成功
	this.sendP2PRes(senderConn, 200, "")
}

// 给发送者返回私聊的结果
func (this *SmsProcess) sendP2PRes(conn net.Conn, code int, errMsg string) {
	var resMes message.Message
	resMes.Type = message.P2PChatResMesType

	var p2pResMes message.P2PChatResMes
	p2pResMes.Code = code
	p2pResMes.Error = errMsg

	data, _ := json.Marshal(p2pResMes)
	resMes.Data = string(data)
	data, _ = json.Marshal(resMes)

	tf := &utils.Transfer{
		Conn: conn,
	}
	err := tf.WritePkg(data)
	if err != nil {
		fmt.Println("sendP2PRes write err:", err)
	}
}

// 转发问价给目标用户
func (this *SmsProcess) SendFileToUser(mes *message.Message, sendConn net.Conn) {
	var fileMes message.FileSendMes
	err := json.Unmarshal([]byte(mes.Data), &fileMes)
	if err != nil {
		fmt.Println("sendFileToUser json.Unmarshal err:", err)
		return
	}
	targetUp, err := userMgr.GetOnlineUserById(fileMes.TargetId)
	if err != nil {
		this.sendFileRes(sendConn, 404, fileMes.FileName, "目标用户不在线")
		fmt.Println("sendFileToUser GetOnlineUserById err:", err)
		return
	}
	data, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("sendFileToUser json.Marshal err:", err)
		return
	}
	//传data给targetUp.Conn
	this.SendMesToEachOnlineUser(data, targetUp.Conn)
	//最后一块通知发送者完成

	if fileMes.ChunkIndex == fileMes.TotalChunk-1 {
		this.sendFileRes(sendConn, 200, fileMes.FileName, "")
	}
}

func (this *SmsProcess) sendFileRes(sendConn net.Conn, code int, fileName string, errMsg string) {
	var resMes message.Message
	resMes.Type = message.FileSendResMesType
	res := message.FileSendResMes{
		Code:     code,
		FileName: fileName,
		Error:    errMsg,
	}
	data, _ := json.Marshal(res)
	resMes.Data = string(data)
	data, _ = json.Marshal(resMes)
	tf := &utils.Transfer{
		Conn: sendConn,
	}
	err := tf.WritePkg(data)
	if err != nil {
		fmt.Println("sendFileRes write err:", err)
		return
	}
}
