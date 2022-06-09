package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

type client chan<- string

var (
	entering = make(chan client) // Событие коннекта
	leaving  = make(chan client) // Событие дисконнекта
	messages = make(chan string) // Сообщения (все)
	gmessage = make(chan string) // Сообщения (от сервера/глобальные)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000") // Подключение
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()   // Горутина событий
	go globalMessage() // Горутина сообщений
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func globalMessage() {
	/**
	* Обработчик ввода в консоли сервера
	**/
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, _ := reader.ReadString('\n')
			text = text[:len(text)-1]
			gmessage <- text
		}
	}()

	for {
		select {
		case msg := <-gmessage:
			messages <- "GLOBAL SERVER MESSAGE: " + msg // Сообщения из консоли сервера
		default:
			messages <- time.Now().Format("15:04:05") // Сообщение с таймером
		}
		time.Sleep(1 * time.Second) // Ждем секунду
	}
}

func handleConn(c net.Conn) {
	ch := make(chan string)
	entering <- ch
	fmt.Println(c.RemoteAddr().String() + " connected")

	/**
	* Отправка сообщений в поток
	**/
	for msg := range ch {
		_, err := fmt.Fprintln(c, msg)
		if err != nil {
			break
		}
	}

	leaving <- ch
	fmt.Println(c.RemoteAddr().String() + " disconnected")
	c.Close()
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages: // Рассылка сообщений по потокам
			for cli := range clients {
				cli <- msg
			}
		case cli := <-entering: // При новом соединении
			clients[cli] = true
		case cli := <-leaving: // При отключении клиента
			delete(clients, cli)
			close(cli)
		}
	}

}
