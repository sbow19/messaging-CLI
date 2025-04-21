package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

type conn struct {
	ws *websocket.Conn

	// Message from the server
	messages chan Response

	// Error
	err chan error

	// Connection closed
	done chan struct{}
}

func NewConnection(ws *websocket.Conn) *conn {
	return &conn{
		ws:       ws,
		messages: make(chan Response),
		done:     make(chan struct{}),
	}
}

// Establish connection with backend and create message channel
func dialBackend() error {
	// Prepare a custom WebSocket config
	origin := "ws://localhost:8000/"
	config, err := websocket.NewConfig(origin, "http://localhost/")
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	// Generate API key to connect with backend
	detailsFile := "details.txt"
	if _, err := os.Stat(detailsFile); err != nil {
		//If file doesn't exist
		if os.IsNotExist(err) {
			// Get random id
			randomId, _ := generateAPIKey()

			apiString := "API_KEY=" + randomId

			// Generate new file with api key
			os.WriteFile(detailsFile, []byte(apiString), 0644)
		} else {
			// Other errors break
			log.Fatal("Failed to read details file")
		}
	}

	// Fetch API key from details file
	key, readErr := ReadAPIKey(detailsFile)
	if readErr != nil {
		log.Fatal("Failed to read details file")
	}

	encoded := base64.StdEncoding.EncodeToString(fmt.Appendf([]byte{}, "%q:", key))
	config.Header = http.Header{
		"Authorization": []string{"Basic " + encoded},
	}

	// Set up initial handshake with server
	conn, err := websocket.DialConfig(config)
	if err != nil {
		log.Fatalf("WebSocket upgrade failed: %v", err)
	}

	// Assign connection to my conn struct
	myconn := NewConnection(conn)

	go myconn.backendListener()

	return nil
}

func (c *conn) listen() {
	for {
		// Receive in responses from backend
		if e := websocket.JSON.Receive(c.ws, c.messages); e != nil {
			c.done <- struct{}{}
			break
		}
	}
}

// Listen for messages from the backend and listen accordingly on goroutine
func (c *conn) backendListener() error {
	defer c.ws.Close()

	go c.listen()

	// Keep readloop active while listening for backend calls
readLoop:
	for {

		select {
		case response := <-c.messages:

			// Determine operation based on message received
			switch response.GetCode() {
			case LoginDetailsRequired, NewLoginDetails:
				// Prompt login details
				promptLoginDetails(ui)
				break
			case LoginSuccessful:
				break
			case IncorrectLogin:
				// Incorrect details
				break
			case Welcome:
				// Display welcome data in mini-message center, terminal input
				break

			}

		case <-c.done:
			break readLoop
		}
	}
	return nil
}

// Send message to the backend
func (c *conn) SendMessage(m *ClientMessage) {
	// Receive in responses from backend
	if e := websocket.JSON.Send(c.ws, m); e != nil {
		c.err <- e
	}
}
