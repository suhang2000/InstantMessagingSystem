package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	// create client object
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	// dial server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error: ", err)
		return nil
	}
	client.conn = conn

	return client
}

func (client *Client) menu() bool {
	var num int
	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")
	_, err := fmt.Scanln(&num)
	if err != nil {
		fmt.Println("请输入0～3的数字")
		return false
	}

	if num >= 0 && num <= 3 {
		client.flag = num
		return true
	} else {
		fmt.Println(">>>>请输入合法范围内的数字<<<<")
		return false
	}
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}

		switch client.flag {
		case 1:
			fmt.Println("public chat")
			client.PublicChat()
		case 2:
			fmt.Println("private chat")
			client.PrivateChat()
		case 3:
			client.UpdateName()
		}
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名<<<<")
	var name string
	_, err := fmt.Scanln(&name)
	if err != nil {
		return false
	}
	client.Name = name

	sendMsg := "rename|" + client.Name + "\n"
	_, err = client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err: ", err)
		return false
	}

	return true
}

func (client *Client) DealResponse() {
	_, err := io.Copy(os.Stdout, client.conn)
	if err != nil {
		return
	}
}

func (client *Client) PublicChat() {
	var msg string
	fmt.Println(">>>>请输入聊天内容，exit退出")
	_, err := fmt.Scanln(&msg)
	if err != nil {
		fmt.Println("输入不合法")
		return
	}

	for msg != "exit" {
		if len(msg) > 0 {
			sendMsg := msg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err: ", err)
				return
			}
		}

		msg = ""
		fmt.Println(">>>>请输入聊天内容，exit退出")
		_, err := fmt.Scanln(&msg)
		if err != nil {
			fmt.Println("输入不合法")
			return
		}
	}
}

func (client *Client) SelectUsers() {
	msg := "who\n"
	_, err := client.conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("conn Write err: ", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	for {
		fmt.Println("当前在线用户：")
		client.SelectUsers()

		fmt.Println(">>>>请输入私聊对象名，exit退出")
		_, err := fmt.Scanln(&remoteName)
		if err != nil {
			fmt.Println("输入用户名错误")
			return
		}

		if remoteName == "exit" {
			break
		}

		for {
			fmt.Println(">>>>请输入私聊内容，exit退出")
			_, err := fmt.Scanln(&chatMsg)
			if err != nil {
				break
			}

			if chatMsg == "exit" {
				break
			}

			if len(chatMsg) > 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err: ", err)
					break
				}
			}
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server IP address")
	flag.IntVar(&serverPort, "port", 8888, "server port")
}

func main() {
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>> dial server failed...")
		return
	}
	fmt.Println(">>>>dial server succeed...")

	go client.DealResponse()

	client.Run()
}
