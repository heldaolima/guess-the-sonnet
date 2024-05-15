package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	PORT              = "9999"
	POEMS_DIR         = "poems/"
	END_OF_CONN_MSG   = "!!!"
	LINES_IN_A_SONNET = 14
)

type Sonnet struct {
	file   os.DirEntry
	title  string
	author string
}

var sonnets []Sonnet
var lenSonnets int

func getLineFromPoem(r io.Reader, lineNum int) (string, error) {
	lastLine := 0
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if lastLine == 0 {
			lastLine++
			continue // skip author - title
		}

		if lastLine == lineNum {
			return scanner.Text(), scanner.Err()
		}

		lastLine++
	}

	return "", io.EOF
}

func getPoems() error {
	files, err := os.ReadDir(POEMS_DIR)
	if err != nil {
		fmt.Println("Error reading poems dir: ", err)
		return err
	}

	lenSonnets = len(files)
	sonnets = make([]Sonnet, lenSonnets)

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
			title:  strings.Trim(metadata[1][:len(metadata[1])-1], " "),
			author: strings.Trim(metadata[0], " "),
		}
	}

	return nil
}

func handlePlay(conn net.Conn) {

	sonnetsCopy := sonnets
	for {
		rand.Shuffle(lenSonnets, func(i, j int) {
			sonnetsCopy[i], sonnetsCopy[j] = sonnetsCopy[j], sonnetsCopy[i]
		})

		randIdx := rand.Intn(lenSonnets)
		sonnet, err := os.ReadFile(POEMS_DIR + sonnetsCopy[randIdx].file.Name())
		if err != nil {
			fmt.Printf("Error openning poem %v: %v", sonnetsCopy[randIdx].file.Name(), err)
			conn.Write([]byte("Erro inicializar jogo"))
			return
		}

		randLine := rand.Intn(LINES_IN_A_SONNET) + 1
		line, err := getLineFromPoem(strings.NewReader(string(sonnet)), randLine)
		if err != nil {
			fmt.Printf("Error getting line from %v: %v", sonnetsCopy[randIdx].file.Name(), err)
			conn.Write([]byte("Erro inicializar jogo"))
			return
		}

		msg := fmt.Sprintf(`-- Jogo dos Sonetos --
- Você receberá um verso de um soneto
- Basta escolher, dentre as alternativas, a qual soneto pertence o verso
------------------------------------------

Verso: %v
------------------------------------------
Opções:`, line)

		for i := 0; i < lenSonnets; i++ {
			msg += fmt.Sprintf("\n[%v] %v", i, sonnets[i].title)
		}

		conn.Write([]byte(msg))

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			client_msg := scanner.Text()

			choice, err := strconv.Atoi(strings.Trim(client_msg, " "))
			if choice == -1 {
				conn.Write([]byte("Saindo do jogo"))
				return
			}

			if err != nil || choice >= lenSonnets || choice < 0 {
				conn.Write([]byte("Mensagem inválida: [" + client_msg + "]"))
				continue
			}

			if choice == randIdx {
				conn.Write([]byte("Resposta correta!"))
				return
			} else {
				conn.Write([]byte("Resposta incorreta."))
				continue
			}

		}
	}
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
		fmt.Printf("(read) Message received: [%v]\n", client_msg)

		choice, err := strconv.Atoi(strings.Trim(client_msg, " "))
		if choice == -1 {
			conn.Write([]byte("Saindo do modo leitura"))
			return
		}
		if err != nil || choice >= lenSonnets || choice < 0 {
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
	msg := fmt.Sprintf("Jogo dos %v sonetos \n\t[1] Ler Sonetos \n\t[2] Jogo dos Sonetos", lenSonnets)

	_, err := conn.Write([]byte(msg))
	if err != nil {
		fmt.Println("Error writing to connection: ", err)
		return
	}

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		client_msg := scanner.Text()
		fmt.Printf("(main) Message received: [%v]\n", client_msg)

		choice, err := strconv.Atoi(strings.Trim(client_msg, " "))
		if err != nil {
			conn.Write([]byte("Mensagem inválida: [" + client_msg + "]"))
		}

		switch choice {
		case 1:
			handleReadPoems(conn)
			// continue
		case 2:
			handlePlay(conn)
			// continue
		case -1:
			fmt.Printf("Connection finished with [%v]\n", conn.RemoteAddr())
			conn.Write([]byte("!!!"))
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

		fmt.Println("Connection accepted [", conn.RemoteAddr(), "]")
		go handler(conn)
	}
}
