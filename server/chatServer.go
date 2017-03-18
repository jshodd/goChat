package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

type ChatRoom struct {
	clientCount int
	allClients  map[net.Conn]int
	newCon      chan net.Conn
	deadCon     chan net.Conn
	messages    chan Message
}

type Message struct {
	Name string
	Text string
}

//adds a new client and listens for messages from the client
func (data *ChatRoom) addConnection(conn net.Conn) {
	data.allClients[conn] = data.clientCount
	data.clientCount = data.clientCount + 1
	log.Print(data.clientCount)
	go func(conn net.Conn) {
		for {
			inbound := new(Message)
			err := json.NewDecoder(conn).Decode(&inbound)
			if err != nil {
				break
			}
			data.messages <- *inbound
		}
		//Disconnecting User
		data.deadCon <- conn

	}(conn)
}

//sends message to all connected clients
func (data *ChatRoom) broadcast(message Message) {
	for conn, _ := range data.allClients {
		go func(conn net.Conn, message Message) {
			err := json.NewEncoder(conn).Encode(message)
			if err != nil {
				data.deadCon <- conn
			}
		}(conn, message)
	}
	log.Printf("New Message: %s", message)
	log.Printf("Broadcast to %d clients", len(data.allClients))
}

//removes dead clients
func (data *ChatRoom) removeConnection(conn net.Conn) {
	log.Printf("Client %d disconnected", data.allClients[conn])
	delete(data.allClients, conn)
}

func main() {
	data := &ChatRoom{
		clientCount: 0,
		allClients:  make(map[net.Conn]int),
		newCon:      make(chan net.Conn),
		deadCon:     make(chan net.Conn),
		messages:    make(chan Message),
	}
	//starting TCP server
	server, err := net.Listen("tcp", ":4040")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Printf("Server Started")

	//go routine to accept clients and loop forever
	//and adds new connections to newCon channel
	go func() {
		for {

			conn, err := server.Accept()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			data.newCon <- conn
			log.Print(data)
		}
	}()

	for {
		select {
		case conn := <-data.newCon:
			data.addConnection(conn)
		case message := <-data.messages:
			data.broadcast(message)
		case conn := <-data.deadCon:
			data.removeConnection(conn)
		}
	}

}
