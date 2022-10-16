package main

func main() {
	newServer := NewServer("127.0.0.1", 8888)
	newServer.Start()
}
