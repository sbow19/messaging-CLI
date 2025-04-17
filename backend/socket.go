package main

import (
	"fmt"
	"io"
	"sync"

	"golang.org/x/net/websocket"
)

// Server factory
type Server struct {
	clients conns
	mu      sync.Mutex
}

type conns map[apiKey]ClientConnection

type ClientConnection struct {
	conn *websocket.Conn
}

func NewServer() *Server {
	return &Server{
		clients: make(map[apiKey]ClientConnection),
		mu:      sync.Mutex{},
	}
}

// Handler multiplexed off to handl individual socket connection
func (s *Server) handleWS(ws *websocket.Conn, k apiKey) {
	s.mu.Lock()
	s.clients[k] = ClientConnection{
		conn: ws,
	}
	s.mu.Unlock()

	// When loop breaks or returns, remove the connection pointer
	defer func() {
		s.mu.Lock()
		delete(s.clients, k)
		s.mu.Unlock()
	}()

	s.readLoop(ws)
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

		ws.Write([]byte("Thank you for the message"))

	}
}
