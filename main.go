package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type Client struct {
	conn   net.Conn
	name   string
	server *Server
}

type Server struct {
	clients  map[net.Conn]*Client
	messages chan string
}

func NewServer() *Server {
	return &Server{
		clients:  make(map[net.Conn]*Client),
		messages: make(chan string),
	}
}

func (s *Server) Broadcast(message string) {
	for _, client := range s.clients {
		fmt.Fprintf(client.conn, "%s\n", message)
	}
}

func (s *Server) Listen() {
	for msg := range s.messages {
		s.Broadcast(msg)
	}
}

func (c *Client) Listen() {
	defer c.conn.Close()
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		msg := scanner.Text()
		c.server.messages <- fmt.Sprintf("[%s:] %s", c.name, msg)
	}

	delete(c.server.clients, c.conn)
	c.server.Broadcast(fmt.Sprintf("%s has left the chat", c.name))
}

func handleConnection(conn net.Conn, server *Server) {
	fmt.Fprintf(conn, "Enter your name: ")
	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	name := scanner.Text()
	client := &Client{
		conn:   conn,
		name:   name,
		server: server,
	}

	server.clients[conn] = client
	server.Broadcast(fmt.Sprintf("%s has joined the chat", name))
	client.Listen()
}

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	server := NewServer()
	go server.Listen()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn, server)
	}
}
