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

	// Receive messages intended for networking part of the app. Subscribe to network slice
	RecNetMess chan *AppMessage

	// Send messages on UI broadcast channel
	UIBroadcast chan *AppMessage

	// Error
	err chan error

	// Connection closed
	done chan struct{}
}

func NewConnection(ws *websocket.Conn, c chan *AppMessage) *conn {
	return &conn{
		ws:          ws,
		RecNetMess:  make(chan *AppMessage),
		UIBroadcast: c,
		messages:    make(chan Response),
		done:        make(chan struct{}),
	}
}

// Establish connection with backend and create message channel
func dialBackend(state *appState) error {
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
	myconn := NewConnection(conn, state.UIBroadcast)

	// Subscribe to UI Messages
	state.SubscribeChannel(myconn.RecNetMess, Network)

	// Listen to messages from network or app
	myconn.listen()

	return nil
}

func (c *conn) listenSocket() {

	for {
		// Receive in Response from backend, then send on
		data := &ClientResponse{
			Message: "Whats up",
		}
		if e := websocket.JSON.Receive(c.ws, data); e != nil {
			c.done <- struct{}{}
			// Close channels on connection object until reestablished
			close(c.messages)
			close(c.done)
			break
		}
		c.messages <- data

	}
}

// Listen for messages from the backend and listen accordingly on goroutine
func (c *conn) listen() error {
	defer c.ws.Close()

	// Listen to websocket messages
	go c.listenSocket()

	// Keep readloop active while listening for backend calls
readLoop:
	for {
		select {
		//Wait for messages from network, and send to UI
		case response := <-c.messages:

			switch response.GetCode() {
			case LoginDetailsRequired:
				// Login details required
				c.UIBroadcast <- &AppMessage{
					Code:    LoginDetailsRequired,
					Message: "Please login",
				}
			case IncorrectLogin:
				// Login details required
				c.UIBroadcast <- &AppMessage{
					Code:    LoginDetailsRequired,
					Message: "Error: login details incorrect",
				}
			}

		// Receive message intended for networking part of app
		case message := <-c.RecNetMess:

			switch message.Code {

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
