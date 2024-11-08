package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

/*
	命令行客户端的实现
*/

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}

	// 链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn

	return client
}

var serverIp string
var serverPort int

func init() {
	// 初始化几个命令
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "请设置服务器IP地址,默认为127.0.0.1")
	flag.IntVar(&serverPort, "port", 8181, "请设置服务器端口,默认为8181")
}

// 循环处理 server的回执消息
func (this *Client) dealResponse() {
	// 不断的读取conn中的消息，返回到控制台，永久阻塞监听
	io.Copy(os.Stdout, this.conn)
	// 等价于 👇

	// for {
	// 	buf := make([]byte, 4096)
	// 	this.conn.Read(buf)
	// 	fmt.Println(buf)
	// }
}

// 菜单展示
func (this *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scan(&flag)

	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>请输入合法范围内的数字<<<<<<")
		return false
	}
}

// 更新用户名
func (this *Client) Rename() bool {
	fmt.Println("请输入用户名...")
	fmt.Scanln(&this.Name)

	sendMsg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// 公聊模式
func (this *Client) PublicWechat() bool {

	var chatMsg string

	fmt.Println(">>>>> 请输入聊天内容,exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 消息不为空，则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				return false
			}
		}
		chatMsg = ""
		fmt.Println(">>>>> 请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)
	}

	return true
}

func (this *Client) SelectUser() {
	sendWoMsg := "who" + "\n"
	_, err := this.conn.Write([]byte(sendWoMsg))
	if err != nil {
		fmt.Println("conn.Write who err:", err)
	}
}

func (this *Client) PrivateWechat() bool {

	this.SelectUser()
	fmt.Println(">>>>> 请选择聊天对象,exit退出")

	var targetUser string
	var sendMsg string
	var writedMsg string

	fmt.Scanln(&targetUser)

	for targetUser != "exit" {

		fmt.Println(">>>>> 请输入聊天内容,exit退出")

		fmt.Scanln(&sendMsg)

		for sendMsg != "exit" {
			// 消息不为空，则发送
			if len(sendMsg) != 0 {
				writedMsg = "to|" + targetUser + "|" + sendMsg + "\n"

				fmt.Println(writedMsg)
				_, err := this.conn.Write([]byte(writedMsg))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					return false
				}
			}

			sendMsg = ""
			fmt.Println(">>>>> 请输入聊天内容,exit退出")
			fmt.Scanln(&sendMsg)
		}
		targetUser = ""
		this.SelectUser()
		fmt.Println(">>>>> 请选择聊天对象,exit退出")
		fmt.Scanln(&targetUser)
	}

	return true
}

func (this *Client) Run() {
	for this.flag != 0 {
		for !this.menu() {
		}

		// 根据不同的模式，处理不同的业务
		switch this.flag {
		case 1:
			// 公聊模式
			this.PublicWechat()
		case 2:
			// 私聊模式
			this.PrivateWechat()
		case 3:
			// 更新用户名
			this.Rename()
		}
	}
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println(">>>>>>>>链接客户端失败...")
		return
	}

	// 单独开启一个goroutine 去处理 server的回执消息
	go client.dealResponse()

	fmt.Println(">>>>>>>>链接客户端成功...")

	// 启动客户端的业务
	client.Run()

}
