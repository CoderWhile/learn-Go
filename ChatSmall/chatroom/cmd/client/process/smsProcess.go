package process

import (
	"ChatSmall/chatroom/cmd/client/utils"
	"ChatSmall/chatroom/cmd/common/message"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SmsProcess struct {
}

// 发送群聊消息
func (this *SmsProcess) SendGroupMes(content string) (err error) {
	//创建一个
	var mes message.Message
	mes.Type = message.SmsMesType

	//SmsMes实例
	var smsMes message.SmsMes
	smsMes.Content = content
	smsMes.UserId = CurUser.UserId
	smsMes.UserStatus = CurUser.UserStatus
	smsMes.UserName = CurUser.UserName

	//3序列化
	data, err := json.Marshal(smsMes)
	if err != nil {
		fmt.Println("SendGroupMes json.Marshal err:", err)
		return
	}
	mes.Data = string(data)
	data, err = json.Marshal(mes)
	if err != nil {
		fmt.Println("SendGroupMes2 json.Marshal err:", err)
		return
	}
	//发送mes
	tf := utils.Transfer{
		Conn: CurUser.Conn,
	}
	err = tf.WritePkg(data)
	if err != nil {
		fmt.Println("SendGroupMes WritePkg err:", err)
		return
	}
	return
}

// 发送私聊消息
func (this *SmsProcess) SendP2PChat(content string, targetId int) (err error) {
	//创建一个Message
	var mes message.Message
	mes.Type = message.P2PChatMesType

	//P2PChatMes实例
	var p2pMes message.P2PChatMes
	p2pMes.Content = content
	p2pMes.TargetId = targetId
	p2pMes.UserId = CurUser.UserId
	p2pMes.UserName = CurUser.UserName
	p2pMes.UserStatus = CurUser.UserStatus

	//序列化p2pMes
	data, err := json.Marshal(p2pMes)
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
		fmt.Println("P2PChat WritePkg err:", err)
		return
	}
	return
}

// 发送文件
func (this *SmsProcess) SendFile(targetId int, filPath string) (err error) {
	fileData, err := os.ReadFile(filPath)
	if err != nil {
		fmt.Println("读取文件失败err:", err)
		return err
	}
	fileName := filepath.Base(filPath)
	fileSize := len(fileData)
	chunkSize := 32 * 1024
	totalChunk := (fileSize + chunkSize - 1) / chunkSize
	//计算块的边界
	for i := 0; i < totalChunk; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > fileSize {
			end = fileSize
		}

		// Base64 编码
		chunkData := base64.StdEncoding.EncodeToString(fileData[start:end])
		var mes message.Message
		mes.Type = message.FileSendMesType
		fileMes := message.FileSendMes{
			FileName:   fileName,
			TargetId:   targetId,
			FileSize:   fileSize,
			TotalChunk: totalChunk,
			ChunkIndex: i,
			Data:       chunkData,
		}
		fileMes.UserId = CurUser.UserId
		fileMes.UserStatus = CurUser.UserStatus
		fileMes.UserName = CurUser.UserName
		data, err := json.Marshal(fileMes)
		if err != nil {
			fmt.Println("SendFileMes json.Marshal err:", err)
			return err
		}
		mes.Data = string(data)
		data, err = json.Marshal(mes)
		if err != nil {
			fmt.Println("SendFileMes json.Marshal err:", err)
			return err
		}
		tf := utils.Transfer{
			Conn: CurUser.Conn,
		}
		err = tf.WritePkg(data)
		if err != nil {
			fmt.Println("SendFileMes WritePkg err:", err)
			return err
		}
		fmt.Println("发送第", i, "块数据")
	}
	return err
}
