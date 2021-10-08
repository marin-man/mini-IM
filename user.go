package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	Server *Server
}

// 创建一个用户的 API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		Server: server,
	}

	// 监听当前 user channel 消息的 goroutine
	go user.ListenMessage()

	return user
}

// 监听当前 User channel 的方法，一旦有消息，就直接发给客户端
func (user *User) ListenMessage() {
	for {
		msg := <-user.C
		user.conn.Write([]byte(msg + "\n"))
	}
}

// 用户上线
func (user *User) Online() {
	// 用户上线，将用户加入到 onLineMap 中
	user.Server.mapLock.Lock()
	user.Server.OnlineMap[user.Name] = user
	user.Server.mapLock.Unlock()
	// 广播当前用户上线消息
	user.Server.BroadCast(user, "已上线")
}

// 用户下线
func (user *User) Offline() {
	// 用户上线，将用户加入到 onLineMap 中
	user.Server.mapLock.Lock()
	delete(user.Server.OnlineMap, user.Name)
	user.Server.mapLock.Unlock()
	// 广播当前用户上线消息
	user.Server.BroadCast(user, "已上线")
}

// 给当前用户发送消息
func (user *User) SendMsg(msg string) {
	user.conn.Write([]byte(msg))
}

// 处理业务
func (user *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前在线用户都有哪些？
		user.Server.mapLock.Lock()
		for _, user := range user.Server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "：" + "在线...\n"
			user.SendMsg(onlineMsg)
		}
		user.Server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 修改用户名
		newName := strings.Split(msg, "|")[1]
		// 判断 name 是否存在
		if _, ok := user.Server.OnlineMap[newName]; ok {
			user.SendMsg("当前用户名被占用\n")
		} else {
			user.Server.mapLock.Lock()
			delete(user.Server.OnlineMap, user.Name)
			user.Server.OnlineMap[newName] = user
			user.Server.mapLock.Unlock()
			user.Name = newName
			user.SendMsg("您已经更新用户名：" + user.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// 消息格式：to|张三|消息内容

		// 1. 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			user.SendMsg("消息格式不正确，请使用 \"to|张三|你好啊\"格式")
			return
		}

		// 2. 根据用户名 得到对方 User 对象
		remoteUser, ok := user.Server.OnlineMap[remoteName]
		if !ok {
			user.SendMsg("该用户名不存在")
			return
		}

		// 3. 获取消息内容，通过对方的
		content := strings.Split(msg, "|")[2]
		if content == "" {
			user.SendMsg("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMsg(user.Name + "对您说：" + content)

	} else {
		user.Server.BroadCast(user, msg)
	}
}
