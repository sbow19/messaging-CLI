package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

type LoginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	origin := "ws://localhost:8000/" // needed by the x/net/websocket client

	// Prepare a custom WebSocket config
	config, err := websocket.NewConfig(origin, "http://localhost/")
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	// Encode username and password for Basic header
	encoded := base64.StdEncoding.EncodeToString([]byte("123456:"))

	config.Header = http.Header{
		"Authorization": []string{"Basic " + encoded},
	}

	// Dial and upgrade the connection
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatalf("WebSocket upgrade failed: %v", err)
	}

	defer ws.Close()

	// 1. Send login credentials
	login := LoginDetails{
		Username: "hello",
		Password: "password",
	}
	if err := websocket.JSON.Send(ws, login); err != nil {
		log.Fatalf("Failed to send login: %v", err)
	}

	// 2. Read login response
	var loginResp string
	if err := websocket.Message.Receive(ws, &loginResp); err != nil {
		log.Fatalf("Failed to receive login response: %v", err)
	}
	fmt.Println("Server:", loginResp)

	if loginResp != "Thank you for the message" {
		fmt.Println("Login failed or not accepted. Exiting.")
		return
	}

	// 3. Message loop (you → server)
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter messages to send (Ctrl+C to quit):")
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}
		text := scanner.Text()

		// Send to server
		if err := websocket.Message.Send(ws, text); err != nil {
			log.Printf("Send error: %v", err)
			break
		}

		// Receive reply -- Note blocking forever if backend crashes
		var reply string
		if err := websocket.Message.Receive(ws, &reply); err != nil {
			log.Printf("Receive error: %v", err)
			break
		}
		fmt.Println("Server:", reply)
	}
}
