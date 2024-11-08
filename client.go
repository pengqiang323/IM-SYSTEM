package main

import (
	"flag"
	"fmt"
	"net"
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

var serverIp string
var serverPort int

func init() {
	// 初始化几个命令
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "请设置服务器IP地址,默认为127.0.0.1")
	flag.IntVar(&serverPort, "port", 8181, "请设置服务器端口,默认为8181")
}

func (this *Client) Run() {
	for this.flag != 0 {
		for !this.menu() {
		}

		// 根据不同的模式，处理不同的业务
		switch this.flag {
		case 1:
			// 公聊模式
			fmt.Println("公聊模式选择...")
		case 2:
			// 私聊模式
			fmt.Println("公聊模式选择...")
		case 3:
			// 更新用户名
			fmt.Println("更新用户名选择...")
		}
	}
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client != nil {
		fmt.Println(">>>>>>>>链接客户端成功...")
		// 启动客户端的业务
	} else {
		fmt.Println(">>>>>>>>链接客户端失败...")
		return
	}

	client.Run()

}
