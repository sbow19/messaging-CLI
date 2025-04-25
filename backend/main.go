package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

/*
	1) Listen to new client connections - DONE
	2) Check whether is new user based on API key provided - DONE
	3) If new, then create credentials, API key and prompt login info - DONE
	4) Establish new Client connection object - DONE

	// EVENTS FROM CLIENT SIDE
	5) Listen to messages and errors from all the client communications
	6) Listen to friend requests, pending, rejected, and accepted
	7) Maintain friends list for each  client. In DB, we have many to many connection, so
		primary key for dual column.
	8) Broadcast new client login/active/inactive to friend list.
	9) Send messages between clients. Messages ae
	10) On client login, synchronise message data.
*/

func main() {
	// go_sqlite3 equired CGO_ENABLED
	os.Setenv("CGO_ENABLED", "1")

	// Load all user data into memory
	err := loadDB()

	if err != nil {
		log.Fatalf("Error loading database: %q", err)
	}

	// Create socket server
	wsServer := NewServer()

	// Main routing handle func
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Spin off new websocket connection handler
		wsServer.start(w, r)
	})

	// Get all users --> prints out to txt file
	go dbConn.GetAllUsers()

	//Listen for app wide messages, e.g. for broadcasting to multiple clients
	com := make(chan *BackendMessage)
	go AppListener(com, wsServer)

	// Start server on PORT
	fmt.Println("HTTP server started at http://localhost:8000")

	// This  appears to be a blocking operation
	listenErr := http.ListenAndServe(":8000", nil)
	if listenErr != nil {
		fmt.Println("Error starting server:", listenErr)
	}

}
