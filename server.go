package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	// online users
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	// broadcast channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}
	// close listen socket
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			fmt.Println("listener close err: ", err)
		}
	}(listener)

	go this.ListenMessager()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err: ", err)
			continue
		}

		go this.Handler(conn)
	}
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("Server Handler: build connection successfully")
	user := NewUser(conn)
	// add user to onlineMap
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()
	// broadcast message
	this.BroadCast(user, "online")

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				// 客户端合法关闭
				this.BroadCast(user, "offline")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read Error: ", err)
				return
			}
			// 提取用户信息，去除末尾'\n'(listenMessage中conn.Write加了'\n')
			msg := string(buf[:n-1])
			this.BroadCast(user, msg)
		}
	}()

	select {}

}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		// send msg to all online users
		this.mapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}
