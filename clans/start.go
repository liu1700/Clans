package clans

import (
	"clans/clans/clan_server"
	"fmt"
)

func Start() {
	server := clan_server.New("127.0.0.1:9090")

	server.OnNewClient(func(c *clan_server.Client) {
		// new client connected
		// lets send some message
		// c.Send("Hello")
		fmt.Println("new client incoming ", c)
	})
	server.OnNewMessage(func(c *clan_server.Client, message string) {
		// new message received
		fmt.Println("received: ", message)
	})
	server.OnClientConnectionClosed(func(c *clan_server.Client, err error) {
		// connection with client lost
		fmt.Println("close ")
	})

	server.Listen()

	return
}
