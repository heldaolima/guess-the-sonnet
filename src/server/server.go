package main

import (
	"bufio"
	"fmt"
	// "io/fs"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	// HOST = "localhost"
	PORT      = "9999"
	POEMS_DIR = "poems/"
)

type Sonnet struct {
	file   os.DirEntry
	title  string
	author string
}

var sonnets []Sonnet

func handlePlay(conn net.Conn) {
	conn.Write([]byte("Will play a game"))
}

func handleReadPoems(conn net.Conn) {
	msg := ""
	for i, poem := range sonnets {
		msg += fmt.Sprintf("[%v] %v - %v\n", i, poem.author, poem.title)
	}

	conn.Write([]byte(msg))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		client_msg := scanner.Text()
		fmt.Printf("Message received: [%v]\n", client_msg)

		choice, err := strconv.Atoi(strings.Trim(client_msg, " "))
		if err != nil || choice >= len(sonnets) || choice < 0 {
			conn.Write([]byte("Mensagem inválida: [" + client_msg + "]"))
			continue
		}

		poem, err := os.ReadFile(POEMS_DIR + sonnets[choice].file.Name())
		if err != nil {
			fmt.Println("Error opening", sonnets[choice].title, ": ", err)
			conn.Write([]byte("Error opening " + sonnets[choice].title))
			continue
		}

		conn.Write(poem)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection: ", err)
		return
	}

}

func handler(conn net.Conn) {
	var msg string
	fmt.Println("Got poems: ", sonnets)

	msg = fmt.Sprintf("Jogo dos %v sonetos \n\t[1] Ler Sonetos \n\t[2] Jogo dos Sonetos", len(sonnets))

	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error writing to connection: ", err)
		return
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		client_msg := scanner.Text()
		fmt.Printf("Message received: [%v]\n", client_msg)

		choice, err := strconv.Atoi(strings.Trim(client_msg, " "))
		if err != nil {
			conn.Write([]byte("Mensagem inválida: [" + client_msg + "]"))
		}

		switch choice {
		case 1:
			handleReadPoems(conn)
			continue
		case 2:
			handlePlay(conn)
			continue
		case -1:
			fmt.Printf("Connection finished with [%v]\n", conn.RemoteAddr())
			return
		default:
			conn.Write([]byte(fmt.Sprintf("Opção inválida: %v", choice)))
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection: ", err)
		return
	}

}

func main() {
	err := getPoems()
	if err != nil {
		fmt.Println("Error getting the poems: ", err)
		return
	}

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

func getPoems() error {
	files, err := os.ReadDir(POEMS_DIR)
	if err != nil {
		fmt.Println("Error reading poems dir: ", err)
		return err
	}

	sonnets = make([]Sonnet, len(files))

	for i, file := range files {
		filePtr, err := os.Open(POEMS_DIR + file.Name())
		if err != nil {
			fmt.Printf("Error opening %v: %v\n", file, err)
			return err
		}
		defer filePtr.Close()

		reader := bufio.NewReader(filePtr)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Printf("Error reading %v: %v\n", file.Name(), err)
			return err
		}

		metadata := strings.Split(string(line), "-")

		sonnets[i] = Sonnet{
			file:   file,
			title:  metadata[1][:len(metadata[1])-1],
			author: metadata[0],
		}
	}

	return nil
}
