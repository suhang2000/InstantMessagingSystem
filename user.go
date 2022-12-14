package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser : create user
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// start listening
	go user.listenMessage()

	return user
}

func (this *User) Online() {
	// add user to onlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	// broadcast message
	this.server.BroadCast(this, "online")
}

func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	this.server.BroadCast(this, "offline")
}

// 处理消息业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		// list online users
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ": online...\n"
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		this.Rename(msg)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// listen message of a user
func (this *User) listenMessage() {
	for {
		msg := <-this.C
		_, err := this.conn.Write([]byte(msg + "\n"))
		if err != nil {
			return
		}
	}
}

func (this *User) SendMsg(msg string) {
	_, err := this.conn.Write([]byte(msg))
	if err != nil {
		return
	}
}

func (this *User) Rename(msg string) {
	// msg: rename|newName
	splits := strings.Split(msg, "|")
	if len(splits) != 2 {
		fmt.Println("rename parameter is illegal")
		return
	}
	newName := splits[1]
	_, ok := this.server.OnlineMap[newName]
	if ok {
		this.SendMsg("this name is already in use\n")
	} else {
		this.server.mapLock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()

		this.Name = newName
		this.SendMsg("rename successfully, new name: " + newName + "\n")
	}
}
