package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) (*Client, error) {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil, err
	}

	client.conn = conn

	client.flag = -1

	return client, nil

}

func (c *Client) ResponseHandler() {
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) menu() bool {
	menu := `
	1. group chat
	2. direct messaging (DM)
	3. update username
	0. exit
	`
	fmt.Println(menu)

	input := readStdin()

	if len(input) == 0 {
		fmt.Println("invalid flag!")
		return false
	}

	flag, err := strconv.Atoi(string(input[0]))

	if err != nil || flag < 0 || flag > 3 {
		fmt.Println("invalid flag!")
		return false
	}

	c.flag = flag
	return true
}

// https://github.com/segmentio/go-prompt/pull/4/files
func readStdin() string {
	msgReader := bufio.NewReader(os.Stdin)
	msgBytes, _, _ := msgReader.ReadLine()
	return string(msgBytes)
}

func (c *Client) UpdateName() bool {
	fmt.Println(">>>>Enter your new username:")
	c.Name = readStdin()

	sendMsg := "{rename}" + c.Name
	_, err := c.conn.Write([]byte(sendMsg + "\n"))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}

	return true

}

func (c *Client) GroupChat() {
	var chatMsg string

	fmt.Println(">>>>Enter your message here (exit to end):")

	chatMsg = readStdin()

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			_, err := c.conn.Write([]byte(chatMsg + "\n"))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				return
			}
		}

		chatMsg = ""
		fmt.Println(">>>>Enter your message here (exit to end):")
		chatMsg = readStdin()

	}

}

func (c *Client) GetOnlineUsers() {
	sendMsg := "who"
	_, err := c.conn.Write([]byte(sendMsg + "\n"))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func (c *Client) DM() {
	c.GetOnlineUsers()
	fmt.Println(">>>>Enter the username you want to talk to (exit to end):")
	remoteName := readStdin()

	for remoteName != "exit" {
		if remoteName == "" {
			continue
		}

		fmt.Println(">>>>Enter your message here (exit to end):")
		chatMsg := readStdin()
		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := fmt.Sprintf("{to=%s}%s", remoteName, chatMsg)
				_, err := c.conn.Write([]byte(sendMsg + "\n"))
				if err != nil {
					fmt.Println("conn.Write err:", err)
					return
				}
			}

			chatMsg = ""
			fmt.Println(">>>>Enter your message here (exit to end):")
			chatMsg = readStdin()
		}

		remoteName = ""
		fmt.Println(">>>>Enter the username you want to talk to (exit to end):")
		remoteName = readStdin()
	}
}

func (c *Client) Run() {
	for c.flag != 0 {
		for !c.menu() {
		}

		switch c.flag {
		case 1:
			fmt.Println("group chat starting...")
			c.GroupChat()
		case 2:
			fmt.Println("DM starting...")
			c.DM()
		case 3:
			fmt.Println("updating username...")
			c.UpdateName()
		}
	}
}

func main() {
	// command line flags
	var serverIp string
	var serverPort int
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "server ip, default 127.0.0.1")
	flag.IntVar(&serverPort, "port", 8888, "server port, default 8888")

	client, err := NewClient(serverIp, serverPort)
	if err != nil {
		fmt.Println("cannot dial server 127.0.0.1:8888")
		return
	}

	go client.ResponseHandler()

	fmt.Println("client connected to 127.0.0.1:8888")

	client.Run()

}
