package main

import (
	"fmt"
	"net"
	"strings"
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

func (this *User) User_SendMsg(msg string) {
	this.Conn.Write([]byte(msg + "\n"))
}

// 消息处理
func (this *User) User_Message(msg string) {
	if msg == "/who" {
		// 查询在线用户
		this.sv.Msglock.Lock()
		for _, u := range this.sv.UserMap {
			this.User_SendMsg("[" + u.Name + "]" + "---在线")
		}
		this.sv.Msglock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "/rename" {
		newName := strings.Split(msg, " ")[1]
		fmt.Println("User changed name,newName:", newName)
		// 查看是否已经有该名字
		this.sv.Msglock.Lock()
		_, ok := this.sv.UserMap[newName]
		if ok {
			this.User_SendMsg("该用户名已存在")
			this.sv.Msglock.Unlock()
			return
		} else {
			// 删除原有用户,map 的 key 不能被修改
			delete(this.sv.UserMap, this.Name)
			this.sv.UserMap[newName] = this
			this.sv.Msglock.Unlock()
			this.Name = newName
			this.User_SendMsg("用户名已更新")
			return
		}
		// 格式 /chat 张三 你好
	} else if len(msg) > 5 && msg[:5] == "/chat" {
		fmt.Println("msg:", msg)
		chatInfo := strings.Split(msg, " ")
		if len(chatInfo) < 3 {
			this.User_SendMsg("格式错误，正确用法: /chat 用户名 内容")
			return
		}
		chatObj := chatInfo[1]
		chatContent := strings.Join(chatInfo[2:], "")
		fmt.Println("User create private chat,object:", chatObj)
		fmt.Println("User create private chat,content:", chatContent)
		this.sv.Msglock.Lock()
		chatUser, hasUser := this.sv.UserMap[chatObj]
		this.sv.Msglock.Unlock()
		if !hasUser {
			this.User_SendMsg("没有该用户")
			return
		}
		chatUser.User_SendMsg("[" + chatObj + "]" + "对你说:" + chatContent)
	} else {
		this.sv.BroadCast(this, msg)
	}
}
