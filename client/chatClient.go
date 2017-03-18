package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

type Message struct {
	Name string
	Text string
}

func encrypt(text string, key []byte) string {
	plaintext := []byte(text)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	return base64.URLEncoding.EncodeToString(ciphertext)
}
func decrypt(cryptoText string, key []byte) string {
	ciphertext, _ := base64.URLEncoding.DecodeString(cryptoText)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return fmt.Sprintf("%s", ciphertext)
}

func main() {

	//Encryption Key
	key := []byte("astaxie12798akljzmknm.ahkjkljl;k")

	//Taking username
	reader := bufio.NewReader(os.Stdin)
	var uName string
	fmt.Println("Enter your username: ")
	uName, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	//Establish connection to server
	conn, err := net.Dial("tcp", ":4040")
	if err != nil {
		fmt.Println("Server unavailable")
		os.Exit(1)
	}
	fmt.Println("Connected!\n")

	//Wait for new messages to send out, infinite loop
	go func() {
		encoder := json.NewEncoder(conn)
		for {
			outbound, _ := reader.ReadString('\n')
			outMessage := Message{
				Name: encrypt(uName, key),
				Text: encrypt(outbound, key),
			}
			encoder.Encode(&outMessage)
		}
	}()

	//Wait for incoming messages, infinite loop
	go func(conn net.Conn) {
		decoder := json.NewDecoder(conn)
		for {
			inMessage := new(Message)
			err := decoder.Decode(&inMessage)
			if err != nil {
				panic(err)
			}
			name := decrypt(inMessage.Name, key)
			msg := decrypt(inMessage.Text, key)
			//if client sent the message, don't print
			if name != uName {
				fmt.Println(name[0:len(name)-1] + ":" + msg[0:len(msg)-1] + "\n")
			}
		}
	}(conn)

	//infinite loop to keep go functions running, probably not the best practice.
	for {
	}
}
