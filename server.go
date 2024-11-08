package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

// 定义Server类
type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User

	// map锁，防止OnlineMap并发
	mapLock sync.RWMutex

	Message chan string
}

// 创建一个Server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 转发消息：监听Mesage广播消息channel的grouting，一旦有消息，就转发消息至全部的在线User
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		fmt.Println("ListenMessager msg = ", msg)

		// 将msg转发给全部在线的User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 发送上线消息提醒到Server
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Name + "]" + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) receiveMsg(user *User) {
	// 创建一个字节的数组，用于接收管道消息
	buf := make([]byte, 4096)

	for {
		// 读取消息,会等待
		fmt.Println("Reading...")
		n, err := user.conn.Read(buf)
		fmt.Println("Read Success...")

		if n == 0 {
			// 用户下线
			user.Offline()
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err=", err)
			return
		}

		// 提取用户的消息，末尾去掉【\n】
		msg := string(buf[:n-1])

		// 将得到的消息，给当前user进行处理
		user.DoMsg(msg)

		// 给心跳管道发送信息
		user.isLive <- true
	}
}

// 当前链接的具体业务
func (this *Server) Handler(conn net.Conn) {

	user := NewUser(conn, this)

	// 用户上线，广播通知
	user.Online()

	// 开启协程，轮巡管道中的消息，给user进行处理
	go this.receiveMsg(user)

	for {
		select {
		case <-user.isLive:
			// 不需要做任何动作，因为select的特性，刷新select，等待下一个case的触发
		case <-time.After(time.Second * 300):
			fmt.Println("触发踢出行为！")
			// 接收一个 10秒后触发的一个管道，
			// 如果case触发成功，则证明10s内，用户无动作。则踢出
			user.conn.Write([]byte("你被踢出了！"))

			// 将user从OnlineMap踢出
			this.mapLock.Lock()
			delete(this.OnlineMap, user.Name)
			this.mapLock.Unlock()

			// 关闭资源
			close(user.C)
			close(user.isLive)
			user.conn.Close()

			// 退出整个Handler
			runtime.Goexit()
		}
	}
}

// 启动服务器的接口
func (this *Server) Start() {
	// 监听socket
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("Listen socket,发生异常，err=", err)
	}

	// 设置一个def，用于方法结束后执行关闭socket
	defer listen.Close()

	// 转发消息
	go this.ListenMessager()

	for {
		// accept 会一直阻塞
		fmt.Println("Accept.....")

		conn, err := listen.Accept()
		fmt.Println("Accept Succes.....")

		if err != nil {
			fmt.Println("Accept 发生异常，err=", err)
			continue
		}

		// 创建一个gorouting协程，相当于给当前用户单独开启了一个管道，for进入下一个循环，等待新的用户加入

		go this.Handler(conn)

	}

}
