package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type Message struct {
	name string
	text string
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	var uName string
	in := make(chan string)
	out := make(chan json)
	fmt.Println("Enter your username: ")
	uName, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}
	conn, err := net.Dial("tcp", ":4040")
	if err != nil {
		fmt.Println("Server unavailable")
		os.Exit(1)
	}
	fmt.Println("Connected!")

	go func() {
		encoder := json.NewEncoder(conn)
		for {
			outbound, _ := reader.ReadString('\n')
			outMessage := Message{
				name: uName,
				text: outbound,
			}
			encoder.Encode(&outMessage)
		}
	}()

	go func(conn net.Conn) {
		decoder := json.NewDecoder(conn)
		for {
			inMessage = new(Message)
			err := decoder.Decode(&inMessage)
			fmt.Println(inMessage.name + " : " + inMessage.text)
		}
	}(conn)

	for {
		select {
		case inMessage := <-in:
			fmt.Println(inMessage)
		case outMessage := <-out:
			_, err = conn.Write([]byte(uName + " : " + outMessage))
		}
	}

}
