package main

import "net"

type User struct {
	Name string
	Addr string      // remote Addr
	Conn net.Conn    // 与客户端的连接
	Chn  chan string // 接受广播
}

func (u *User) User_Listener() {
	for {
		// channel没数据回阻塞接受，一有数据就发送给客户端
		msg := <-u.Chn
		u.Conn.Write([]byte(msg + "\n"))
	}
}

func AddUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	newUser := &User{
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		Chn:  make(chan string),
	}
	// 启动监听当前user channel的goroutine
	go newUser.User_Listener()
	return newUser
}
