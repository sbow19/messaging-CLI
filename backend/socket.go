package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

// Server factory
type Server struct {
	clients conns
	mu      sync.Mutex
}

type conns map[apiKey]*ClientConnection

type ClientConnection struct {
	conn *websocket.Conn
}

func (c *ClientConnection) SendOnConnection(m Response) *RequestError {

	jsonData, err := json.Marshal(m)

	if err != nil {
		return &RequestError{
			Message: "Failed to send message",
			Code:    FailedMessageSend,
		}
	}

	// Attempt to write. If failed then close websocket connection
	_, e := c.conn.Write(jsonData)

	if e != nil {

		return &RequestError{
			Message: "Failed to send message",
			Code:    FailedMessageSend,
		}

	}

	return nil

}

func NewServer() *Server {
	return &Server{
		clients: make(map[apiKey]*ClientConnection),
		mu:      sync.Mutex{},
	}
}

// Start websocket connection
func (s *Server) start(w http.ResponseWriter, r *http.Request) {

	websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			// Close websocket connection and remove from connections map
			ws.Close()
		}()

		// Get API key from frontend
		k, err := getApiKey(r)

		if err != nil {
			websocket.JSON.Send(ws, err)
			return
		}

		// Check if api key exists
		doesUserExist(k)

		// Set new client connection in server clients map
		s.setConnection(ws, k)

		s.clients[k].SendOnConnection(
			&ClientResponse{
				Err:     nil,
				Message: "API Key Existing",
				Code:    APIKey,
			})

		s.handleWS(ws, k)

	}).ServeHTTP(w, r)

}

func (s *Server) setConnection(ws *websocket.Conn, k apiKey) {
	s.mu.Lock()
	s.clients[k] = &ClientConnection{
		conn: ws,
	}
	s.mu.Unlock()
}

// Handler multiplexed off to handl individual socket connection
func (s *Server) handleWS(ws *websocket.Conn, k apiKey) {

	// When loop breaks or returns, remove the connection pointer
	defer func() {
		s.mu.Lock()
		delete(s.clients, k)
		s.mu.Unlock()

		// Log user out
		UserMap[k].Leave()
	}()

	// Send welcome message if not sent
	if !UserMap[k].welcomeSent {
		err := s.clients[k].SendOnConnection(
			&ClientResponse{
				Err:     nil,
				Message: "Welcome to the server!",
				Code:    Welcome,
			})
		if err != nil {
			// Tear down connection by returning from handler
			return
		}

		UserMap[k].welcomeSent = true
	}

	// Send prompt for login details
	authErr := s.authLoop(ws, k)
	if authErr != nil {
		s.clients[k].SendOnConnection(
			&ClientResponse{
				Err:     authErr,
				Message: authErr.Message,
				Code:    authErr.Code,
			})

		// Clean up connection with the client
		return
	}

	// Start listening to frontend messages
	s.readLoop(ws)
}

// Communicate regarding authentication
func (s *Server) authLoop(ws *websocket.Conn, k apiKey) *RequestError {

	newMessage := &AuthResponse{
		Code:    LoginDetailsRequired,
		Message: "Login details required",
	}
	err := s.clients[k].SendOnConnection(newMessage)
	if err != nil {
		return &RequestError{
			Message: "Error while logging in",
			Code:    AuthenticationError,
		}
	}

	// Send and receive auth details and responses
	for {
		var resp LoginDetails
		err := websocket.JSON.Receive(ws, &resp)

		if err != nil {
			return &RequestError{
				Message: "Connection error",
				Code:    ConnectionError,
			}
		}

		// Authenticate new client
		auth, authE := authenticationCycle(k, &resp)

		if authE != nil {
			return &RequestError{
				Message: "Error while logging in",
				Code:    AuthenticationError,
			}
		}

		if auth.Code == LoginSuccessful {
			// Resend auth message
			err := s.clients[k].SendOnConnection(auth)

			if err != nil {
				return &RequestError{
					Message: "Connection error",
					Code:    ConnectionError,
				}
			}
			return nil
		} else {
			// Resend auth message
			err := s.clients[k].SendOnConnection(auth)

			if err != nil {
				return &RequestError{
					Message: "Connection error",
					Code:    ConnectionError,
				}
			}
		}
	}

}

// Read incoming messages from the clients
func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {

			if err == io.EOF {
				break
			}

			fmt.Println("Read error: ", err)
			continue
		}
		msg := buf[:n]
		fmt.Println(string(msg))

		ws.Write(buf)

	}
}
