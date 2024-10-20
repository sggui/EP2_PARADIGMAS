package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type Client struct {
	Nick     string
	Channel  chan string
	IsBot    bool
	Conn     net.Conn
	Outgoing chan<- Message
}

type Message struct {
	Sender  *Client
	Content string
	Private bool
	Target  string
}

var (
	entering = make(chan *Client)
	leaving  = make(chan *Client)
	messages = make(chan Message)
)

func broadcaster() {
	clients := make(map[*Client]bool)
	nicknames := make(map[string]*Client)
	for {
		select {
		case msg := <-messages:
			if msg.Private {
				targetClient, exists := nicknames[msg.Target]
				if exists {
					targetClient.Channel <- fmt.Sprintf("@%s enviou no privado: %s", msg.Sender.Nick, msg.Content)
					fmt.Printf("Mensagem privada para @%s to @%s: %s\n", msg.Sender.Nick, targetClient.Nick, msg.Content)
				} else {
					msg.Sender.Channel <- fmt.Sprintf("Usuário @%s não encontrado.", msg.Target)
				}
			} else {
				for cli := range clients {
					if !cli.IsBot {
						cli.Channel <- fmt.Sprintf("@%s escreveu: %s", msg.Sender.Nick, msg.Content)
					}
				}
			}

		case cli := <-entering:
			clients[cli] = true
			nicknames[cli.Nick] = cli
			announcement := fmt.Sprintf("%s @%s acabou de entrar", func() string {
				if cli.IsBot {
					return "Bot"
				}
				return "NULL"
			}(), cli.Nick)
			for c := range clients {
				c.Channel <- announcement
			}
			fmt.Println(announcement)

		case cli := <-leaving:
			delete(clients, cli)
			delete(nicknames, cli.Nick)
			close(cli.Channel)
			announcement := fmt.Sprintf("@%s saiu", cli.Nick)
			for c := range clients {
				c.Channel <- announcement
			}
			fmt.Println(announcement)
		}
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	input := bufio.NewScanner(conn)
	var nick string
	var isBot bool

	conn.Write([]byte("Escreva seu nome de exibição: "))
	if input.Scan() {
		nick = strings.TrimSpace(input.Text())
	}

	conn.Write([]byte("voce é o bot? (sim/não): "))
	if input.Scan() {
		response := strings.TrimSpace(strings.ToLower(input.Text()))
		if response == "sim" || response == "s" {
			isBot = true
		}
	}

	ch := make(chan string)
	client := &Client{
		Nick:     nick,
		Channel:  ch,
		IsBot:    isBot,
		Conn:     conn,
		Outgoing: messages,
	}

	go clientWriter(client)

	entering <- client

	nickChange := func(newNick string) {
		oldNick := client.Nick
		client.Nick = newNick
		messages <- Message{
			Sender:  client,
			Content: fmt.Sprintf("Usuário @%s mudou o nome para @%s", oldNick, newNick),
			Private: false,
		}
	}

	for input.Scan() {
		text := input.Text()
		if strings.HasPrefix(text, "\\") {
			args := strings.SplitN(text, " ", 3)
			switch args[0] {
			case "\\changenick":
				if len(args) >= 2 {
					nickChange(args[1])
				} else {
					client.Channel <- "Usando: \\changenick [new_nickname]"
				}
			case "\\msg":
				if len(args) >= 3 {
					targetNick := strings.TrimPrefix(args[1], "@")
					messages <- Message{
						Sender:  client,
						Content: args[2],
						Private: true,
						Target:  targetNick,
					}
				} else {
					client.Channel <- "Usando: \\msg [@nickname] [message]"
				}
			case "\\exit":
				return
			default:
				client.Channel <- "Comando não encontrado."
			}
		} else {
			messages <- Message{
				Sender:  client,
				Content: text,
				Private: false,
			}
		}
	}

	leaving <- client
}

func clientWriter(client *Client) {
	for msg := range client.Channel {
		fmt.Fprintln(client.Conn, msg)
	}
}

func main() {
	fmt.Println("Startando o server...")
	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}
