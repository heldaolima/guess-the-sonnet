package main

import (
	"bufio"
	"fmt"
	"guess_the_sonnet_server/sonnets"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	PORT              = "9999"
	POEMS_DIR         = "sonnets/poems/"
	END_OF_CONN_MSG   = "!!!"
	LINES_IN_A_SONNET = 14
)

var poems []sonnets.Sonnet
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

func handlePlay(conn net.Conn) {
	poemsCopy := poems
	for {
		rand.Shuffle(lenSonnets, func(i, j int) {
			poemsCopy[i], poemsCopy[j] = poemsCopy[j], poemsCopy[i]
		})

		randIdx := rand.Intn(lenSonnets)
		randLine := rand.Intn(LINES_IN_A_SONNET) + 1

		line, err := poemsCopy[randIdx].GetLine(POEMS_DIR, randLine)
		if err != nil {
			fmt.Printf("Error getting line from %v: %v", poemsCopy[randIdx].File.Name(), err)
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
			msg += fmt.Sprintf("\n[%v] %v", i, poemsCopy[i].Title)
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
	for i, poem := range poems {
		msg += fmt.Sprintf("[%v] %v - %v\n", i, poem.Author, poem.Title)
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

		poem, err := os.ReadFile(POEMS_DIR + poems[choice].File.Name())
		if err != nil {
			fmt.Println("Error opening", poems[choice].Title, ": ", err)
			conn.Write([]byte("Error opening " + poems[choice].Title))
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
	var err error

	poems, err = sonnets.GetSonnets(POEMS_DIR)
	if err != nil {
		fmt.Println("Error getting the poems: ", err)
		return
	}

	lenSonnets = len(poems)

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
