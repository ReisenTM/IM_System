package main

import (
	"fmt"
	"net"
)

type User struct {
	Name string
	Addr string      // remote Addr
	Conn net.Conn    // 与客户端的连接
	Chn  chan string // 接受广播
	sv   *Server     // 当前属于哪个服务器
}

// 用户channel监听
func (u *User) User_Listener() {
	for {
		// channel没数据回阻塞接受，一有数据就发送给客户端
		msg := <-u.Chn
		// 写入数据到客户端
		_, err := u.Conn.Write([]byte(msg + "\n"))
		if err != nil {
			fmt.Println("[User_Listener] Write error", err)
			return
		}
	}
}

// 增加新用户
func AddUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	newUser := &User{
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		Chn:  make(chan string),
		sv:   server,
	}
	// 启动监听当前user channel的goroutine
	go newUser.User_Listener()
	return newUser
}

// 用户上线
func (this *User) User_Online() {
	this.sv.Msglock.Lock()
	this.sv.UserMap[this.Name] = this
	this.sv.Msglock.Unlock()
	this.sv.BroadCast(this, "用户已上线")
}

// 用户下线
func (this *User) User_Offline() {
	this.sv.Msglock.Lock()
	delete(this.sv.UserMap, this.Name)
	this.sv.Msglock.Unlock()
	this.sv.BroadCast(this, "用户已下线")
}

// 消息处理
func (this *User) User_Message(msg string) {
	this.sv.BroadCast(this, msg)
}
