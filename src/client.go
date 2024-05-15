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

	if err != nil {
		fmt.Println("Error establishing connection to the server:", err)
		os.Exit(0)
	}

	defer conn.Close()

	fmt.Println("Connection established with the server!")

	for {
		buffer := make([]byte, 1024)
		msgLen, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading content from the server: ", err)
			continue
		}

		serverMsg := string(buffer[:msgLen])
		fmt.Print(serverMsg)
		if serverMsg[:3] == "!!!" {
			fmt.Println("Conexion finished")
			break
		}

		fmt.Print("\nSend: ")
		reader := bufio.NewReader(os.Stdin)
		msg, _ := reader.ReadString('\n')

		_, err = conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Error writing to the server: ", err)
			continue
		}
	}
}
