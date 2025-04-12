package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	mode       int8
}

// 创建新连接
func NewClient(ip string, port int) *Client {
	client := &Client{
		ServerIP:   ip,
		ServerPort: port,
		mode:       127,
	}
	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if conn == nil {
		fmt.Println("创建连接失败", err)
		return nil
	}
	client.conn = conn

	return client
}

// 客户端操作界面
func (c *Client) ShowMenu() bool {
	var flag int8
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.修改名字")
	fmt.Println("0.退出客户端")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		c.mode = flag
		return true
	} else {
		fmt.Println("请输入合法范围内的数字")
		return false
	}
}

// 处理服务器返回的结果
func (c *Client) DealResonse() {
	// 一旦client.conn有数据，就直接copy到stdout，永远阻塞监听
	io.Copy(os.Stdout, c.conn)
}

// 模式
type Mode int8

const (
	Exit Mode = iota
	PublicChat
	PrivateChat
	Rename
)

func (c *Client) DoRename() bool {
	fmt.Print("请输入修改后的名字:")
	fmt.Scanln(&c.Name)
	msg := fmt.Sprintf("/rename %s\n", c.Name)
	_, err := c.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn:Write error", err)
		return false
	}
	return true
}

func Clear() {
	// 清屏函数
	fmt.Print("\033[2J\033[H")
}

func (c *Client) DoPublicChat() {
	fmt.Println("您已进入公聊,请输入内容")
	reader := bufio.NewReader(os.Stdin)
	for {
		// 读到换行为止，防止空格被吞,包括delim
		msg, _ := reader.ReadString('\n')
		if len(msg) >= 5 && strings.Contains(msg, "/quit") {
			break
		} else {
			c.conn.Write([]byte(msg))
		}
	}
}

func (c *Client) DoPrivateChat() bool {
	fmt.Println("您已进入私聊,")
	fmt.Println("请先选择聊天对象")
	query := "/who"
	c.conn.Write([]byte(query + "\n"))
	var toChat string
	fmt.Scanln(&toChat)
	if toChat == "" {
		return false
	}
	fmt.Println("请输入私聊内容")
	reader := bufio.NewReader(os.Stdin)
	msg, _ := reader.ReadString('\n')
	cmd := fmt.Sprintf("%s %s %s", "/chat", toChat, msg)
	c.conn.Write([]byte(cmd))
	return true
}

func (c *Client) Run() {
	for c.mode != int8(Exit) {
		Clear()
		for !c.ShowMenu() {
			// 菜单选择直到正确为止
		}
		switch c.mode {
		case int8(PublicChat):
			fmt.Println("公聊模式>>")
			c.DoPublicChat()
		case int8(PrivateChat):
			fmt.Println("私聊模式>>")
			err := c.DoPrivateChat()
			if !err {
				fmt.Println("请输入正确内容")
			} else {
				fmt.Println("发送成功")
			}
		case int8(Rename):
			fmt.Println("修改名字模式")
			c.DoRename()
		}
	}
}

var (
	ServerIP   string
	ServerPort int
)

func init() {
	// 在main函数之前调用
	// 命令行解析绑定 -> flag包
	flag.StringVar(&ServerIP, "ip", "127.0.0.1", "设置服务器ip地址(默认127.0.0.1)")
	flag.IntVar(&ServerPort, "port", 9999, "设置服务器端口(默认9999)")
}

func main() {
	flag.Parse()
	client := NewClient(ServerIP, ServerPort)
	if client == nil {
		fmt.Println("<-链接服务器失败->")
		return
	}
	go client.DealResonse()
	fmt.Println("《链接服务器成功》")
	client.Run()
}
