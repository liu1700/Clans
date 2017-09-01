package clan_server

import (
	"fmt"
	"log"
	"net"
)

// Client holds info about connection
type Client struct {
	conn   *net.TCPConn
	Server *server
}

// TCP server
type server struct {
	address                  string // Address to open connection: localhost:9999
	onNewClientCallback      func(c *Client)
	onClientConnectionClosed func(c *Client, err error)
	onNewMessage             func(c *Client, message string)
}

// Read client data from channel
func (c *Client) listen() {
	// reader := bufio.NewReader(c.conn)
	// bufio.
	for {
		buf := make([]byte, 4096)
		n, err := c.conn.Read(buf)
		// message, err := reader.ReadString('\n')
		if err != nil {
			c.conn.Close()
			c.Server.onClientConnectionClosed(c, err)
			return
		}
		println("value ", n)
		for _, v := range buf[:n-1] {
			fmt.Println("raw ", v, "string ", string(v))
		}
		// c.Server.onNewMessage(c, message)
		c.Server.onNewMessage(c, string(buf[:n-1]))
	}
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) Conn() *net.TCPConn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

// Called right after server starts listening new client
func (s *server) OnNewClient(callback func(c *Client)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *server) OnClientConnectionClosed(callback func(c *Client, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *server) OnNewMessage(callback func(c *Client, message string)) {
	s.onNewMessage = callback
}

// Start network server
func (s *server) Listen() {
	addr, _ := net.ResolveTCPAddr("tcp", s.address)
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal("Error starting TCP server.")
	}
	defer listener.Close()

	for {
		conn, _ := listener.AcceptTCP()
		client := &Client{
			conn:   conn,
			Server: s,
		}
		go client.listen()
		s.onNewClientCallback(client)
	}
}

// Creates new tcp server instance
func New(address string) *server {
	log.Println("Creating server with address", address)
	server := &server{
		address: address,
	}

	server.OnNewClient(func(c *Client) {})
	server.OnNewMessage(func(c *Client, message string) {})
	server.OnClientConnectionClosed(func(c *Client, err error) {})

	return server
}
