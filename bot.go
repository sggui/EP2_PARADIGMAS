package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
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
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	// Ler e descartar o prompt "Enter your nickname: "
	if scanner.Scan() {
		// Prompt recebido, não precisamos fazer nada aqui
	}

	// Enviar o apelido do bot
	nick := "ReverseBot"
	fmt.Fprintln(conn, nick)

	// Ler e descartar o prompt "Are you a bot? (yes/no): "
	if scanner.Scan() {
		// Prompt recebido, não precisamos fazer nada aqui
	}

	// Indicar que este é um bot
	fmt.Fprintln(conn, "yes")

	// Agora podemos entrar no loop principal para receber mensagens
	for scanner.Scan() {
		msg := scanner.Text()
		fmt.Printf("Received: %s\n", msg)

		// Verificar se a mensagem é direcionada ao bot
		if strings.Contains(msg, "@"+nick) {
			// Extrair a mensagem
			parts := strings.SplitN(msg, ":", 2)
			if len(parts) == 2 {
				originalMsg := strings.TrimSpace(parts[1])

				// Inverter a mensagem
				reversedMsg := reverse(originalMsg)

				// Obter o apelido do remetente
				senderParts := strings.Split(parts[0], "@")
				if len(senderParts) >= 2 {
					senderNick := strings.TrimSpace(senderParts[1])

					// Enviar a resposta privada
					response := fmt.Sprintf("\\msg @%s %s", senderNick, reversedMsg)
					fmt.Fprintln(conn, response)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Erro ao ler do servidor:", err)
	}
}
