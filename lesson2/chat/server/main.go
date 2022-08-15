package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

type client chan<- string

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	question string
	answer   string // Да, здесь больше подошел бы числовой тип данных. Но, так больше вариаций для творчества (можно делать полноценную викторину не меняя весь код)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	go game() // Игра
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
func game() {
	// Реализация игры
	// answer = "5"
	for {
		if len(answer) < 1 {
			question = "Решите задачу: "
			operators := []string{"+", "-", "/", "*"}
			operand1 := rand.Intn(9)
			operand2 := rand.Intn(9)
			operator := operators[rand.Intn(len(operators))]
			if operand2 == 0 && operator == "/" {
				operand2 = 1
			}
			op := map[string]func(int, int) int{
				"+": func(o1, o2 int) int { return o1 + o2 },
				"-": func(o1, o2 int) int { return o1 - o2 },
				"/": func(o1, o2 int) int { return int(o1 / o2) },
				"*": func(o1, o2 int) int { return o1 * o2 },
			}
			question += fmt.Sprintf("%d %s %d = ? ", operand1, operator, operand2)
			answer = fmt.Sprint(op[operator](operand1, operand2))

		}
		messages <- "Математика на скорость: " + question
		time.Sleep(5 * time.Second)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)
	// who := conn.RemoteAddr().String()

	ch <- "Your name: "
	nameBufScan := bufio.NewScanner(conn)
	nameBufScan.Scan()
	clientName := nameBufScan.Text()

	ch <- "You are " + clientName
	messages <- clientName + " has arrived"
	entering <- ch
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg := input.Text()
		if msg == answer {
			messages <- clientName + " is winner! Right answer: " + answer
			answer = ""
		} else {
			messages <- clientName + ": " + input.Text()
		}
	}
	leaving <- ch
	messages <- clientName + " has left"
	conn.Close()
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
