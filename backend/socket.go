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

	// DEbugging with message
	fmt.Println(m)

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
		if !doesUserExist(k) {
			// Add new user details to UserMap
			clientData := generateNewUser(k)

			// Save to database
			err := dbConn.CreateNewUser(clientData)

			// TODO: Parse db error messages
			if err != nil {
				websocket.JSON.Send(ws, &RequestError{
					Message: "Error creating new user",
					Code:    DatabaseError,
				})
				return
			}
			UserMap[k] = clientData
		}

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

	var err *RequestError
	// When loop breaks or returns, remove the connection pointer
	defer func() {
		s.mu.Lock()
		delete(s.clients, k)
		s.mu.Unlock()

		// Log user out
		UserMap[k].Leave()

		// TODO: broadcast to friends that user logged out
	}()

	// Send welcome message if not sent
	if !UserMap[k].welcomeSent {
		err = s.clients[k].SendOnConnection(
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

		// Update database
		dbErr := dbConn.UpdateClient(UserMap[k])
		if dbErr != nil {
			err = s.clients[k].SendOnConnection(
				&ClientResponse{
					Err:     nil,
					Message: "Error saving client data!",
					Code:    DatabaseError,
				})
			if err != nil {
				// Tear down connection by returning from handler
				return
			}

			return
		}
	}

	// Send prompt for login details
	err = s.authLoop(ws, k)
	if err != nil {
		s.clients[k].SendOnConnection(
			&ClientResponse{
				Err:     err,
				Message: err.Message,
				Code:    err.Code,
			})

		// Clean up connection with the client
		return
	}

	// Start listening to frontend messages
	s.readLoop(ws)
}

// Communicate regarding authentication
func (s *Server) authLoop(ws *websocket.Conn, k apiKey) *RequestError {

	var reqErr *RequestError
	var authResp *AuthResponse
	var resp LoginDetails
	var err error

	authResp = &AuthResponse{
		Code:    LoginDetailsRequired,
		Message: "Login details required",
	}

	// Request login details from client
	reqErr = s.clients[k].SendOnConnection(authResp)
	if reqErr != nil {
		goto reqErrSend
	}

	// Send and receive auth details and responses
	for {
		err = websocket.JSON.Receive(ws, &resp)
		if err != nil {
			reqErr = &RequestError{
				Message: "Connection error",
				Code:    ConnectionError,
			}
			goto reqErrSend
		}

		// Authenticate new client
		authResp, reqErr = authenticationCycle(k, &resp)

		if reqErr != nil {
			goto reqErrSend
		}

		if authResp.Code == LoginSuccessful {
			// Resend auth message
			err := s.clients[k].SendOnConnection(authResp)

			if err != nil {
				reqErr = &RequestError{
					Message: "Connection error",
					Code:    ConnectionError,
				}
				goto reqErrSend
			}
			return nil
		} else {
			// Resend auth message
			err := s.clients[k].SendOnConnection(authResp)

			if err != nil {
				return &RequestError{
					Message: "Connection error",
					Code:    ConnectionError,
				}
			}

			// Listen to  user response again
			continue
		}
	}

reqErrSend:
	fmt.Println(err)
	return reqErr
}

// Read incoming messages from the client. Several different operations - friends find, friend request
// and actual text sent between users.
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
