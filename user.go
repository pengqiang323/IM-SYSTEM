package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string // 用来 接收消息 的管道
	isLive chan bool   // 用来处理 在线状态检测 的管道
	conn   net.Conn
	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		isLive: make(chan bool),
		conn:   conn,
		server: server,
	}

	// 启动监听当前【User Channel】消息的 gorouting
	go user.ListenMessage()

	return &user
}

// 用户上线通知
func (this *User) Online() {
	// 用户上线，将用户加入到OnlineMap中（修改OnlineMap时，需要加锁）
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前消息
	this.server.BroadCast(this, "已上线，快来打招呼吧!")
}

// 用户下线通知
func (this *User) Offline() {
	this.conn.Write([]byte("已下线！"))
	this.server.BroadCast(this, "已下线！")
}

// 发送上线消息提醒到Server
func (this *User) getOnlineUsers() {

	userSlice := make([]string, 0)

	for _, cli := range this.server.OnlineMap {
		if cli.Name != this.Name {
			// 将OnlineMap中的用户封装成数组返回出去
			userSlice = append(userSlice, cli.Name+"在线!")
		} else {
			continue
		}
	}

	// 将消息转换为byte数组返回
	writeMsg := []byte(strings.Join(userSlice, "\n"))
	this.conn.Write(writeMsg)
}

func (this *User) Rename(newName string) {

	// 修改用户map时，需锁住
	this.server.mapLock.Lock()

	// 判断新名称，是否已存在
	_, ok := this.server.OnlineMap[newName]
	if ok {
		this.conn.Write([]byte("名称已经被使用！"))
	} else {
		// 删除原user
		delete(this.server.OnlineMap, this.Name)

		// 新增一个user
		this.Name = newName
		this.server.OnlineMap[newName] = this
		this.conn.Write([]byte("名称更新成功！"))
	}
	this.server.mapLock.Unlock()
}

func (this *User) PrivateChat(userName, msg string) {
	// 判断私聊的对象是否存在
	targetUser, ok := this.server.OnlineMap[userName]
	if !ok {
		this.conn.Write([]byte("用户不存在"))
	} else {
		// 新增一个user
		targetUser.C <- msg
	}
}

// 用户处理消息业务
func (this *User) DoMsg(msg string) {
	switch {
	case msg == "who":
		// 查询当前在线用户
		this.getOnlineUsers()
	case strings.Contains(msg, "rename|"):
		// 更改名称
		// 消息格式 rename|张三
		this.Rename(strings.Split(msg, "|")[1])
	case strings.Contains(msg, "to|"):
		// 私聊
		// 消息格式 to|张三|消息内容
		msgArray := strings.Split(msg, "|")
		this.PrivateChat(msgArray[1], msgArray[2])
	default:
		// 广播
		this.server.BroadCast(this, msg)
	}
}

// 监听当前Uer channel 的方法，一旦消息，就直接发送给 端客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
