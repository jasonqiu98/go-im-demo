package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// online users
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// channel for broadcasting
	Broadcast chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Broadcast: make(chan string),
	}

	return server
}

func (s *Server) BroadcastListener() {
	for {
		msg := <-s.Broadcast

		fmt.Println(msg)

		// send msg to all online users
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.ch <- msg
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) ConnectionHandler(conn net.Conn) {
	// create a new user
	user := NewUser(conn, s, make(chan bool))
	user.Online()

	for {
		select {
		case status := <-user.alive:
			if !status {
				// close channel and return
				close(user.alive)
				return
			} else {
				// do nothing / reset the timer
				continue
			}
		case <-time.After(time.Minute * 5): // a timeout of five minutes
			user.Kickout()
			return
		}
	}
}

// the "main" function of the IM Server
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
	}
	// defer close
	defer listener.Close()

	fmt.Printf("listening on %s:%d\n", s.Ip, s.Port)

	go s.BroadcastListener()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// do handler
		go s.ConnectionHandler(conn)
	}

}
