package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net"
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
		err:         make(chan error),
		messages:    make(chan Response),
		done:        make(chan struct{}),
	}
}

// Establish connection with backend and create message channel
func dialBackend(state *appState, addr net.Addr) {
	// Prepare a custom WebSocket config
	var ip string
	switch v := addr.(type) {
	case *net.TCPAddr:
		ip = v.IP.String() // Extract the IP address
	case *net.UDPAddr:
		ip = v.IP.String() // Extract the IP address
	default:
		log.Fatalf("Unsupported address type: %T", v)
	}

	// Manually assign the port (e.g., 8000)
	port := 8000

	// Prepare a custom WebSocket config
	origin := fmt.Sprintf("ws://%s:%d/", ip, port)
	config, err := websocket.NewConfig(
		origin,
		fmt.Sprintf("http://%s:%d/", ip, port)) // Use the correct format
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
		// Send UI message
		log.Println(err)
		aMess := AppMessage{
			Code:    ConnectionError,
			Message: "Error connecting to server",
			Payload: nil,
		}
		state.UIBroadcast <- &aMess
		return

	}

	// Assign connection to my conn struct
	myconn := NewConnection(conn, state.UIBroadcast)

	// Subscribe to UI Messages
	state.SubscribeChannel(myconn.RecNetMess, Network)

	// Listen to messages from network or app
	myconn.listen(state)

}

func (c *conn) listenSocket() {

	for {
		// Receive in Response from backend, then send on
		data := &ClientResponse{}
		if e := websocket.JSON.Receive(c.ws, data); e != nil {
			// If  socket is closed and finish message sent from backend,trigger connection error
			c.done <- struct{}{}
			// Close channels on connection object until reestablished
			close(c.messages)
			close(c.done)
			break
		}
		c.messages <- data

	}
}

// Listen for messages from the frontend and backend
func (c *conn) listen(state *appState) error {
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
			case LoginSuccessful:
				state.SetLoggedIn()
				c.UIBroadcast <- &AppMessage{
					Code:    LoginSuccessful,
					Message: "You are logged in",
				}

			case AllContent:
				/*
					Receive all conetnt from backend.
						1) Friends (and status) - DONE
						2) Friend requests		- DONE
						3) Messages from past few days - DONE
				*/

				var userContent UserContent
				var err error
				err = response.DecodePayload(&userContent)

				if err != nil {

				}
				err = state.AssignAllContent(&userContent)

				if err != nil {

				}

				c.UIBroadcast <- &AppMessage{
					Code:    AllContent,
					Message: "All user content fetched",
				}

			case UpdateFriendContent:
				var userContent UserContent
				var err error
				err = response.DecodePayload(&userContent)

				if err != nil {
					log.Fatal(err)
				}
				err = state.AssignFriendshipContent(&userContent)

				if err != nil {
					log.Fatal(err)

				}

				c.UIBroadcast <- &AppMessage{
					Code:    UpdateFriendContent,
					Message: "Friend data updated",
					Payload: nil,
				}

			case SearchUsersResults:

				c.UIBroadcast <- &AppMessage{
					Code:    SearchUsersResults,
					Message: "Results",
					Payload: response.GetPayload(),
				}
			case FriendRequestResult:
				c.UIBroadcast <- &AppMessage{
					Code:    FriendRequestResult,
					Message: "Results",
					Payload: response.GetPayload(),
				}
			case FriendAcceptResult:
				c.UIBroadcast <- &AppMessage{
					Code:    FriendAcceptResult,
					Message: "Results",
					Payload: response.GetPayload(),
				}

			case ReceiveMessage:
				var message Message
				var err error
				err = response.DecodePayload(&message)

				if err != nil {
					log.Fatal(err)
				}
				err = state.AppendMessage(&message)

				if err != nil {
					log.Fatal(err)

				}

				appMessage := AppMessage{
					Code:    ReceiveMessage,
					Message: "Friend data updated",
					Payload: nil,
				}

				appMessage.EncodePayload(&message)

				c.UIBroadcast <- &appMessage

			case NotifyLogin:
				var user string
				var err error
				err = response.DecodePayload(&user)

				if err != nil {
					log.Fatal(err)

				}

				// Update active status globally
				err = state.SetFriendActiveStatus(
					user,
					true,
				)

				if err != nil {
					log.Fatal(err)

				}

				appMessage := AppMessage{
					Code:    NotifyLogin,
					Message: "User logged in",
					Payload: nil,
				}

				appMessage.EncodePayload(user)

				c.UIBroadcast <- &appMessage
			case NotifyInactive:
				var user string
				var err error
				err = response.DecodePayload(&user)

				if err != nil {
					log.Fatal(err)

				}

				// Update active status globally
				err = state.SetFriendActiveStatus(
					user,
					false,
				)

				if err != nil {
					log.Fatal(err)

				}

				appMessage := AppMessage{
					Code:    NotifyInactive,
					Message: "User logged in",
					Payload: nil,
				}

				appMessage.EncodePayload(user)

				c.UIBroadcast <- &appMessage

			default:

			}

		// Receive message intended for networking part of app
		case message := <-c.RecNetMess:

			switch message.Code {

			case AttemptLogin:
				// Message
				clientMess := ClientMessage{
					Code:    message.Code,
					Payload: message.Payload,
				}
				// Send message
				c.SendMessage(&clientMess)
			case SearchUsers:
				// Message
				clientMess := ClientMessage{
					Code:    message.Code,
					Payload: message.Payload,
				}
				// Send message
				c.SendMessage(&clientMess)
			case FriendRequest:
				// Message
				clientMess := ClientMessage{
					Code:    message.Code,
					Payload: message.Payload,
				}
				// Send message
				c.SendMessage(&clientMess)
			case FriendAccept:
				// Message
				clientMess := ClientMessage{
					Code:    message.Code,
					Payload: message.Payload,
				}
				// Send message
				c.SendMessage(&clientMess)
			case SendMessage:
				// Message
				clientMess := ClientMessage{
					Code:    message.Code,
					Payload: message.Payload,
				}

				// Send message
				c.SendMessage(&clientMess)

			}

		case <-c.done:
			// Broadcast Connection error
			aMess := AppMessage{
				Code:    ConnectionError,
				Message: "Lost connection to server",
			}
			c.UIBroadcast <- &aMess

			// Unsubscribe listener channel
			state.UnsubscribeChannel(c.RecNetMess, Network)
			break readLoop
		}
	}
	return nil
}

// Send message to the backend
func (c *conn) SendMessage(m *ClientMessage) {

	// Receive in responses from backend
	if e := websocket.JSON.Send(c.ws, m); e != nil {
		c.done <- struct{}{}
	}
}

func ListenForBroadcast(myAppState *appState) {

	pc, err := net.ListenPacket("udp4", ":9999")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	buf := make([]byte, 1024)
	for {

		//
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		log.Printf("%s sent this: %s\n", addr, buf[:n])

		go func() {
			// Blocks until connection is lost, then retry takes place
			dialBackend(myAppState, addr)
		}()

		// End loop. Read li
		<-myAppState.done
	}

}
