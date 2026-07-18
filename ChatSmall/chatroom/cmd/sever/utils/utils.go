package utils

import (
	"ChatSmall/chatroom/cmd/common/message"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// 这里 将这些方法关联到结构体中
type Transfer struct {
	//连接，缓冲
	Conn net.Conn
	Buf  [8096]byte //传输时的缓冲
}

// 信息传输
// 读取信息到返回值mes中
func (this *Transfer) ReadPkg() (mes message.Message, err error) {
	//buf := make([]byte, 1024)
	fmt.Println("读取客户端发送的数据。。")
	_, err = io.ReadFull(this.Conn, this.Buf[:4])
	//_, err = this.Conn.Read(this.Buf[:4])
	if err != nil {
		fmt.Println("conn.Read err:", err)
		//err = errors.New("read pkg header error")
		return
	}
	//根据读到的buf[:4}转成unit32
	var pkgLen uint32
	pkgLen = binary.BigEndian.Uint32(this.Buf[:4])

	//根据读到的字节切片大小动态创建buf大小
	buf := make([]byte, pkgLen)
	_, err = io.ReadFull(this.Conn, buf)
	if err != nil {
		return
	}

	//把pkglen反序列化--message.Message
	err = json.Unmarshal(buf, &mes)
	if err != nil {
		fmt.Println("json.Unmarshal err:", err)
		return
	}
	return
}

// 写信息将一个字节切片发送给对方
func (this *Transfer) WritePkg(data []byte) (err error) {
	//1.先发送一个长度给对方
	var pkgLen uint32
	pkgLen = uint32(len(data))
	//	var buf [4]byte
	binary.BigEndian.PutUint32(this.Buf[0:4], pkgLen)
	//发送长度

	n, err := this.Conn.Write(this.Buf[:4])
	if n != 4 || err != nil {
		fmt.Println("conn.Write(bytes) err:", err)
		return
	}
	//2.发送消息本身
	n, err = this.Conn.Write(data)
	if n != int(pkgLen) || err != nil {
		fmt.Println("conn.Write(bytes) err:", err)
		return
	}
	return
}
