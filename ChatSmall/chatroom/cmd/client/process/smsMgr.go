package process

import (
	"ChatSmall/chatroom/cmd/common/message"
	"encoding/json"
	"fmt"
)

func outputGroupMes(mes *message.Message) {
	//1.反序列化
	var smsMes message.SmsMes
	err := json.Unmarshal([]byte(mes.Data), &smsMes)
	if err != nil {
		fmt.Println("json.Unmarshal err:", err)
		return
	}
	//显示信息
	info := fmt.Sprintf("%s Say:： %s", smsMes.UserName, smsMes.Content)
	fmt.Println(info)
}
