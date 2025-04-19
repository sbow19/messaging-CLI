package main

import (
	"fmt"
	"net/http"
)

/*
	1) Listen to new client connections - DONE
	2) Check whether is new user based on API key provided
	3) If new, then create credentials, API key and prompt login info
	4) Establish new Client connection object

	// EVENTS FROM CLIENT SIDE
	5) Listen to messages and errors from all the client communications
	6) Listen to friend requests, pending, rejected, and accepted
	7) Maintain friends list for each  client. In DB, we have many to many connection, so
		primary key for dual column.
	8) Broadcast new client login/active/inactive to friend list.
	9) Send messages between clients. Messages ae
	10) On client login, synchronise message data.
	11) We could combine mongo db and SQL here. Monog t o store message data, linked with a particular
*/

func main() {
	// Create socket server
	wsServer := NewServer()

	// Main routing handle func
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Spin off new websocket connection handler
		wsServer.start(w, r)
	})

	// Start server on PORT
	fmt.Println("HTTP server started at http://localhost:8000")

	// This  appears to be a blocking operation
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
