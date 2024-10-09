package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Get nickname
	fmt.Print("Informe seu nome de exibição: ")
	reader := bufio.NewReader(os.Stdin)
	nick, _ := reader.ReadString('\n')
	nick = strings.TrimSpace(nick)
	conn.Write([]byte(nick + "\n"))

	// Indicate that this is not a bot
	conn.Write([]byte("no\n"))

	// Channels to handle I/O
	done := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		close(done)
	}()

	// Read input from stdin and send to server
	inputScanner := bufio.NewScanner(os.Stdin)
	for inputScanner.Scan() {
		text := inputScanner.Text()
		if text == "\\exit" {
			break
		}
		fmt.Fprintln(conn, text)
	}

	conn.Close()
	<-done
}
