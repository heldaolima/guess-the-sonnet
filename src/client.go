package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "9999"
	SERVER_TYPE = "tcp"
)

func main() {
	conn, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	defer conn.Close()

	if err != nil {
		fmt.Println("Error establishing connection to the server")
		panic(err)
	}

	fmt.Println("Connection established with the server!")

	for {
		fmt.Print("\nSend a message to the server: ")

		reader := bufio.NewReader(os.Stdin)
		msg, _ := reader.ReadString('\n')

		_, err = conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Error writing to the server: ", err)
			return
		}

		buffer := make([]byte, 1024)
		msgLen, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading content from the server: ", err)
			return
		}

		fmt.Println("Received:", string(buffer[:msgLen]))

		if msg == "-1\n" {
			break
		}
	}
}
