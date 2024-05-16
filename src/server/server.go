package main

import (
	"bufio"
	"fmt"
	"guess_the_sonnet_server/sonnets"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	STATE_FIRST_SCREEN = iota
	STATE_RETURN_TO_FIRST_SCREEN
	STATE_SHOW_POEMS
	STATE_SELECT_POEMS
	STATE_POEM_SELECTED
	STATE_START_GAME
	STATE_INVALID_CHOICE
	STATE_BREAK_CONNECTION
	STATE_ERROR
)

const (
	PORT              = "9999"
	POEMS_DIR         = "sonnets/poems/"
	END_OF_CONN_MSG   = "!!!"
	LINES_IN_A_SONNET = 14
)

var poems []sonnets.Sonnet
var previousState int
var lenSonnets int
var chosenSonnetToRead int

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

func getSelectedPoem() (string, error) {
	poem, err := os.ReadFile(POEMS_DIR + poems[chosenSonnetToRead].File.Name())
	if err != nil {
		return "", err
	}

	return string(poem), nil
}

func handleChoice(choice int, state *int) {
	previousState = *state
	switch *state {
	case STATE_FIRST_SCREEN:
		switch choice {
		case -1:
			*state = STATE_BREAK_CONNECTION
		case 1:
			*state = STATE_SHOW_POEMS
		case 2:
			*state = STATE_START_GAME
		default:
			*state = STATE_INVALID_CHOICE
		}
	case STATE_SELECT_POEMS:
		if choice == -1 {
			*state = STATE_FIRST_SCREEN
		} else if choice < -1 || choice > lenSonnets {
			*state = STATE_INVALID_CHOICE
		} else {
			*state = STATE_POEM_SELECTED
			chosenSonnetToRead = choice
		}
	default:
		return
	}
}

func getMsg(state *int) string {
	switch *state {
	case STATE_FIRST_SCREEN:
		return fmt.Sprintf("Jogo dos %v sonetos \n\t[1] Ler Sonetos \n\t[2] Jogo dos Sonetos\n\t[-1] Encerrar conexão", lenSonnets)
	case STATE_SHOW_POEMS:
		var msg string
		for i, poem := range poems {
			msg += fmt.Sprintf("[%v] %v - %v\n", i, poem.Author, poem.Title)
		}
		msg += "[-1] Sair\n"
		*state = STATE_SELECT_POEMS
		return msg

	case STATE_POEM_SELECTED:
		poem, err := getSelectedPoem()
		if err != nil {
			fmt.Println("Error reading selected poem:", err)
			return "Erro ao carregar o poema."
		}
		*state = STATE_RETURN_TO_FIRST_SCREEN
		return poem

	case STATE_BREAK_CONNECTION:
		return "!!!"

	case STATE_INVALID_CHOICE:
		*state = previousState
		return "Mensagem inválida"

	default:
		return ""
	}
}

func handler(conn net.Conn) {
	state := STATE_FIRST_SCREEN

	for {
		msg := getMsg(&state)
		_, err := conn.Write([]byte(msg))
		if err != nil {
			fmt.Println("Error writing to connection: ", err)
			return
		}

		if state == STATE_RETURN_TO_FIRST_SCREEN {
			state = STATE_FIRST_SCREEN
			continue
		}

		if state == STATE_BREAK_CONNECTION {
			fmt.Printf("Connection finished with [%v]\n", conn.RemoteAddr())
			return
		}

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			client_msg := scanner.Text()
			fmt.Printf("Message received: [%v]\n", client_msg)

			choice, err := strconv.Atoi(strings.Trim(client_msg, " "))
			if err != nil {
				conn.Write([]byte("Mensagem inválida: [" + client_msg + "]"))
				continue
			}

			handleChoice(choice, &state)
			break
		}

		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading from connection: ", err)
			return
		}

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
