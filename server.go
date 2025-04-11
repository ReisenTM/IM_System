package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip      string
	Port    int
	UserMap map[string]*User // 存放所有用户
	Message chan string      // 广播管道
	Msglock sync.RWMutex
}

// 服务器初始化
func Server_init(ip string, port int) *Server {
	res := &Server{
		Ip:      ip,
		Port:    port,
		UserMap: make(map[string]*User),
		Message: make(chan string),
	}
	return res
}

func (sv *Server) Server_MsgListener() {
	for {
		msg := <-sv.Message
		// 将msg发送给全部在线user
		sv.Msglock.Lock()
		for _, user := range sv.UserMap {
			user.Chn <- msg
		}
		sv.Msglock.Unlock()
	}
}

// 广播消息的方法
func (sv *Server) BroadCast(user *User, msg string) {
	broadCast := fmt.Sprintf("[地址:%s]", user.Addr)
	sv.Message <- broadCast + msg
}

// 服务器事件处理
func (sv *Server) Server_Handler(conn net.Conn) {
	// fmt.Println("服务器启动")
	user := AddUser(conn, sv)
	user.User_Online()
	go func() {
		// 处理用户输入
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				// 用户下线
				user.User_Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Read error", conn)
				return
			}
			// 对获取到的消息进行提取
			msg := string(buf[:n-1])
			// 对msg进行处理
			user.User_Message(msg)
		}
	}()

	// 永久阻塞
	select {}
}

func (sv *Server) Server_Start() {
	addr := fmt.Sprintf("%s:%d", sv.Ip, sv.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Listen err", err)
		return
	}
	defer listener.Close()
	// 监听message
	go sv.Server_MsgListener()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err", err)
			continue
		}
		// 当有新用户上线时
		go sv.Server_Handler(conn)
	}
}
