package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

const (
	// HOST = "localhost"
	PORT = "9999"
)

func handler(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		client_msg := scanner.Text()
		fmt.Printf("Message received: [%v]\n", client_msg)

		if client_msg == "-1" {
			fmt.Printf("Connection finished with [%v]\n", conn.RemoteAddr())
			conn.Write([]byte("Connection finished"))
			return
		}

		_, err := conn.Write([]byte(strings.ToUpper(client_msg)))
		if err != nil {
			fmt.Println("Error writing to connection: ", err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection: ", err)
		return
	}

}

func main() {
	server, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		fmt.Println("Error creating server:", err)
		return
	}

	defer server.Close()

	fmt.Println("Server listening on port " + PORT)
	for {
		conn, err := server.Accept()
		defer conn.Close()

		if err != nil {
			fmt.Println("Error accepting connection: ", err)
			continue
		}

		fmt.Println("Connection accepetd [", conn.RemoteAddr(), "]")
		go handler(conn)
	}

}
