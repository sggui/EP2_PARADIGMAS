package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func reverse(s string) string {
	runes := []rune(s)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		j := n - i - 1
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func main() {
	for {
		conn, err := net.Dial("tcp", "localhost:3000")
		if err != nil {
			log.Println("Erro ao conectar ao servidor:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		nick := "ReverseBot"
		fmt.Fprintln(conn, nick)
		fmt.Println("Enviado nickname:", nick)

		fmt.Fprintln(conn, "sim")
		fmt.Println("Enviado resposta: sim")

		done := make(chan struct{})

		go func() {
			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				msg := scanner.Text()
				fmt.Printf("Recebi: %s\n", msg)

				
				if strings.Contains(msg, "enviou no privado:") {
					parts := strings.SplitN(msg, "enviou no privado:", 2)
					if len(parts) == 2 {
						senderInfo := strings.TrimSpace(parts[0])
						content := strings.TrimSpace(parts[1])
						senderParts := strings.Split(senderInfo, "@")
						if len(senderParts) == 2 {
							senderNick := strings.TrimSpace(senderParts[1])

							reversedMsg := reverse(content)

							response := fmt.Sprintf("\\msg @%s %s", senderNick, reversedMsg)
							fmt.Fprintln(conn, response)
							fmt.Println("Enviado resposta:", response)
						}
					}
				}
			}
			if err := scanner.Err(); err != nil {
				log.Println("Erro na leitura do servidor:", err)
			}
			close(done)
		}()

		<-done

		conn.Close()
		time.Sleep(5 * time.Second)
	}
}
