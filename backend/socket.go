package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// Server factory
type Server struct {
	clients   conns
	broadcast chan *BackendMessage
	mu        sync.Mutex
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
		clients:   make(map[apiKey]*ClientConnection),
		broadcast: make(chan *BackendMessage),
		mu:        sync.Mutex{},
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

	// Assuming auth loop passed, then get user data
	err = s.SendAllContent(ws, k)
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
	s.readLoop(ws, k)
}

// Communicate regarding authentication
func (s *Server) authLoop(ws *websocket.Conn, k apiKey) *RequestError {

	var reqErr *RequestError
	var authResp *AuthResponse
	var resp ClientMessage
	var loginDetails LoginDetails
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

		// Continue with auth loop. Client cannot send any other type of message
		if resp.Code != AttemptLogin {
			continue
		}

		// Get login details from payload
		err := json.Unmarshal(resp.Payload, &loginDetails)

		if err != nil {
			fmt.Println(err)
		}

		// Authenticate new client
		authResp, reqErr = authenticationCycle(k, &loginDetails)

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

func (s *Server) SendAllContent(ws *websocket.Conn, k apiKey) *RequestError {

	var reqErr *RequestError
	var contentResp *ClientResponse
	var allContent *UserContent
	var err error

	contentResp = &ClientResponse{
		Code:    AllContent,
		Message: "All user content",
		Err:     nil,
		Payload: nil,
	}

	// Get all user content
	allContent, err = dbConn.GetAllUserContent(k)

	if err != nil {
		goto reqErrSend
	}

	//Encode data in client response
	contentResp = &ClientResponse{
		Code:    AllContent,
		Err:     nil,
		Message: "All user content",
		Payload: nil,
	}

	err = contentResp.EncodePayload(allContent)

	if err != nil {
		goto reqErrSend
	}

	err = s.clients[k].SendOnConnection(contentResp)
	if err != nil {
		goto reqErrSend
	}

	return nil

reqErrSend:
	fmt.Println(err)
	return reqErr
}

// Read incoming messages from the client. Several different operations - friends find, friend request
// and actual text sent between users.

type ChatBroadcast struct {
	Chat       *Message  `json:"chat"`
	Friendship *[]string `json:"friendship"`
}

func (s *Server) readLoop(ws *websocket.Conn, k apiKey) {

	var clientMessage ClientMessage
	for {
		err := websocket.JSON.Receive(ws, &clientMessage)

		if err != nil {

			if err == io.EOF {
				break
			}

			fmt.Println("Read error: ", err)
			continue
		}

		switch clientMessage.Code {

		case SearchUsers:
			// Attempt to search database for users
			var srch string
			var err error
			var results *UsersSearch

			clientMessage.DecodePayload(&srch)
			results, err = UserSearchResults(srch)

			if err != nil {
				log.Fatal(err)
			}

			clientResponse := ClientResponse{
				Code:    SearchUsersResults,
				Payload: nil,
				Err:     nil,
				Message: fmt.Sprintf("There were %d results", len(*results)),
			}
			clientResponse.EncodePayload(results)

			err = websocket.JSON.Send(ws, &clientResponse)

			if err != nil {
				log.Fatal(err)
			}

		case FriendRequest:
			// Attempt to search database for users
			var name string // receiver of request
			var err error
			var result string
			var friendRequestId string

			clientMessage.DecodePayload(&name)
			friendRequestId, err = SetFriendRequest(name, string(k))

			if err != nil {
				result = "Failed to save friend request"
			} else {
				result = "Friend request sent"
			}

			clientResponse := ClientResponse{
				Code:    FriendRequestResult,
				Payload: nil,
				Err:     nil,
				Message: "",
			}
			clientResponse.EncodePayload(&result)

			err = websocket.JSON.Send(ws, &clientResponse)

			if err != nil {
				log.Fatal(err)
			}

			// Network broadcast to update clients
			s.broadcast <- &BackendMessage{
				Code:    BroadcastFriendRequest,
				Payload: friendRequestId,
			}

		case FriendAccept:
			var friendAcceptData FriendAcceptData
			var friendIds *[]string
			var err error
			var result string

			clientMessage.DecodePayload(&friendAcceptData)
			friendIds, err = dbConn.GetFriendRequestById(friendAcceptData.RequestId)

			if err != nil {
				log.Fatalln(err)
			}
			err = UpdateFriendRequest(&friendAcceptData, string(k))

			if err != nil {
				result = "Failed to accept request"
			} else {
				result = "Friend accepted successfully"
			}

			clientResponse := ClientResponse{
				Code:    FriendAcceptResult,
				Payload: nil,
				Err:     nil,
				Message: "",
			}
			clientResponse.EncodePayload(&result)

			err = websocket.JSON.Send(ws, &clientResponse)

			if err != nil {
				log.Fatal(err)
			}

			// Network broadcast to update friends under  given friendship ID
			s.broadcast <- &BackendMessage{
				Code:    BroadcastFriendship,
				Payload: friendIds,
			}

		case SendMessage:
			var chat Chat
			var message Message
			var err error
			var friendship *[]string
			// Decode chat message
			err = clientMessage.DecodePayload(&chat)

			if err != nil {
				log.Fatal(err)
			}

			if chat.Receiver == "" {
				break
			}

			// Save message in database
			friendship, err = dbConn.SaveMessage(&chat, k)

			layout := "2006-01-02 15:04"
			nowUTC := time.Now().UTC()
			formatted := nowUTC.Format(layout)

			message = Message{
				Text:     chat.Text,
				Date:     formatted,
				Receiver: chat.Receiver,
				Sender:   chat.Sender,
			}

			if err != nil {
				log.Fatal(err)
			}

			// If receiving user is active, then send new message immediately
			// Network broadcast to update friends under  given friendship ID
			s.broadcast <- &BackendMessage{
				Code: BroadcastChat,
				Payload: &ChatBroadcast{
					Chat:       &message,
					Friendship: friendship,
				},
			}

		}

	}
}

func UserSearchResults(s string) (*UsersSearch, error) {
	//Search db for users
	return dbConn.GetUsers(s)

}

func SetFriendRequest(name string, id string) (string, error) {
	return dbConn.SetFriendRequest(name, id)
}

func UpdateFriendRequest(f *FriendAcceptData, id string) error {
	if f.Accept {
		// Returns new friendship Id
		return dbConn.CreateFriend(f, id)
	} else {
		// Return friend request id
		return dbConn.DeleteFriendRequest(f.RequestId)

	}
}
