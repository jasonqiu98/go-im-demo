package main

import (
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
)

type User struct {
	Name string
	Addr string

	// to receive msg
	ch   chan string
	conn net.Conn

	// server
	srv *Server

	alive chan bool
}

// create a new user
func NewUser(conn net.Conn, server *Server, alive chan bool) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:  userAddr,
		Addr:  userAddr,
		ch:    make(chan string),
		conn:  conn,
		srv:   server,
		alive: alive,
	}

	go user.MessageListener()

	return user
}

// msg formatter for a certain user
func (u *User) MessageWrapper(msg string) string {
	return fmt.Sprintf("[%s]%s:%s", u.Addr, u.Name, msg)
}

// this buffer receives messages from clients
func MessageBuffer(u *User) {
	buf := make([]byte, 1024*4) // 4KB
	for {
		n, err := u.conn.Read(buf)
		if n == 0 {
			// if the channel is not yet closed
			// meaning the user is alive
			if !IsClosed(u.ch) {
				u.Offline(true)
			}
			return
		}

		if err != nil && err != io.EOF {
			fmt.Println("Conn Read err:", err)
			return
		}

		// remove the ending "\n"
		msg := string(buf[:n-1])

		// handle the message from clients
		u.MessageHandler(msg)

		u.alive <- true
	}
}

func (u *User) Online() {
	s := u.srv

	// make user online
	s.mapLock.Lock()
	s.OnlineMap[u.Name] = u
	s.mapLock.Unlock()

	// create a buffer for the user
	go MessageBuffer(u)

	// broadcast "user online"
	s.Broadcast <- u.MessageWrapper("online")
}

// https://go101.org/article/channel-closing.html

func IsClosed[T any](ch <-chan T) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func (u *User) Offline(normal bool) {
	s := u.srv

	// user offline
	s.mapLock.Lock()
	delete(s.OnlineMap, u.Name)
	s.mapLock.Unlock()

	if normal {
		s.Broadcast <- u.MessageWrapper("offline")
	} else {
		s.Broadcast <- u.MessageWrapper("kicked")
	}

	// let MessageListener() to help close u.conn
	if !IsClosed(u.ch) {
		close(u.ch)
	}

	if !IsClosed(u.alive) {
		u.alive <- false
	}
}

func (u *User) Kickout() {
	u.ch <- u.MessageWrapper("user inactive, kicked out...")
	u.Offline(false)
}

// handle the message from clients
func (u *User) MessageHandler(msg string) {
	s := u.srv

	if msg == "" {
		return
	} else if msg == "who" {
		// check who's online
		// format: "who"
		s.mapLock.Lock()
		for _, user := range s.OnlineMap {
			u.ch <- user.MessageWrapper("online...")
		}
		s.mapLock.Unlock()
		return
	} else if strings.HasPrefix(msg, "{rename}") {
		// rename to a new username
		// format: "{rename}..."
		newName := strings.Replace(msg, "{rename}", "", 1)
		_, ok := s.OnlineMap[newName]
		if ok {
			u.ch <- fmt.Sprintf("username %s is already used", newName)
		} else {
			s.mapLock.Lock()
			delete(s.OnlineMap, u.Name)
			s.OnlineMap[newName] = u
			s.mapLock.Unlock()

			oldName := u.Name
			u.Name = newName
			u.ch <- u.MessageWrapper(
				fmt.Sprintf("username changed from %s to %s", oldName, newName),
			)
		}
	} else if r := regexp.MustCompile("^{to=([a-zA-Z0-9]+)}(.+)"); r.MatchString(msg) {
		// DM: direct message
		// format: "{to={receiver}}..."
		parts := r.FindStringSubmatch(msg)
		remoteName, msgBody := parts[1], parts[2]
		remoteUser, ok := s.OnlineMap[remoteName]
		if !ok {
			u.ch <- u.MessageWrapper(
				fmt.Sprintf("username %s does not exist", remoteName),
			)
			return
		}
		remoteUser.ch <- u.MessageWrapper(msgBody)
	} else {
		s.Broadcast <- u.MessageWrapper(msg)
	}
}

// listen to the "user" channel
// the channel is operated on by the server
func (u *User) MessageListener() {
	// for range channel
	for msg := range u.ch {
		_, err := u.conn.Write([]byte(msg + "\n"))
		if err != nil {
			panic(err)
		}
	}

	// close conn
	err := u.conn.Close()
	if err != nil {
		panic(err)
	}
}
