package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

// NewUser : create user
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// start listening
	go user.listenMessage()

	return user
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
