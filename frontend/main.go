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

func main() {
	origin := "ws://localhost:8000/" // needed by the x/net/websocket client

	// Prepare a custom WebSocket config
	config, err := websocket.NewConfig(origin, "http://localhost/")
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	// Get random id
	randomId, _ := generateAPIKey()

	encoded := base64.StdEncoding.EncodeToString(fmt.Appendf([]byte{}, "%q:", randomId))

	config.Header = http.Header{
		"Authorization": []string{"Basic " + encoded},
	}

	// Set up initial handshake with server
	ws, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatalf("WebSocket upgrade failed: %v", err)
	}

	defer ws.Close()

	// Listen to initial handshakes
	var authReply AuthResponse

handshake:
	for {
		if e := websocket.JSON.Receive(ws, &authReply); e != nil {
			log.Fatal(e)
		}

		switch authReply.Code {
		case LoginDetailsRequired, NewLoginDetails:
			break handshake
		}

	}

	// Create new input buffer scanner
	scanner := bufio.NewScanner(os.Stdin)

authLoop:
	for {

		switch authReply.Code {
		case LoginSuccessful:
			break authLoop
		case IncorrectLogin:
			fmt.Println("Incorrect login details")
			fmt.Println("Enter username (Ctrl+C to quit):")
			scanner.Scan()
			username := scanner.Text()
			fmt.Println("Enter password (Ctrl+C to quit):")
			scanner.Scan()
			password := scanner.Text()

			// Send details to backend
			login := &LoginDetails{
				Username: username,
				Password: password,
			}

			err := websocket.JSON.Send(ws, login)
			if err != nil {
				ws.Close()
				log.Fatal("Error sending login details")
			}
		case NewLoginDetails, LoginDetailsRequired:
			fmt.Println("Enter login details")

			fmt.Println("Enter username (Ctrl+C to quit):")
			scanner.Scan()

			username := scanner.Text()

			fmt.Println("Enter password (Ctrl+C to quit):")
			scanner.Scan()

			password := scanner.Text()

			// Send details to backend
			login := &LoginDetails{
				Username: username,
				Password: password,
			}

			err := websocket.JSON.Send(ws, login)
			if err != nil {
				ws.Close()
				log.Fatal("Error sending login details")
			}
		}

		// Wait for auth loop details
		if err := websocket.JSON.Receive(ws, &authReply); err != nil {

			fmt.Println(authReply)
			log.Fatalf("Receive error: %q", err)
		}
	}

chat:
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break chat
		}
		text := scanner.Text()

		// Send to server
		if err := websocket.Message.Send(ws, text); err != nil {
			log.Printf("Send error: %v", err)
			break chat
		}

		// Receive reply -- Note blocking forever if backend crashes
		var reply string
		if err := websocket.Message.Receive(ws, &reply); err != nil {
			log.Printf("Receive error: %v", err)
			break chat
		}
		fmt.Println("Server:", reply)
	}

}
