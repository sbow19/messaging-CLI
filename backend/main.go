package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/websocket"
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

	// Create main context
	ctx := context.Background()

	// Create socket server
	wsServer := NewServer()

	// Main routing handle func
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Server side timeout for ctx
		deadline := time.Now().Add(10 * time.Second)
		ctx, cancel := context.WithDeadline(ctx, deadline)

		// Release context resources
		defer cancel()

		// Finished request channel
		res := make(chan ClientResponse)
		// Spin off handler go routine
		go parseIncomingReq(ctx, res, r)

		for {
			select {
			case d := <-res:

				if d.Err.Code == LoginSuccessful {
					websocket.Handler(func(ws *websocket.Conn) {
						defer func() {
							ws.Close()
						}()

						// TODO: send in timeout context
						wsServer.handleWS(ws, apiKey(d.Message))

					}).ServeHTTP(w, r)
					return
				} else {

					// Marshal the struct to JSON
					jsonData, err := json.Marshal(d)
					if err != nil {
						log.Fatal("Error marshaling JSON:", err)
					}

					reader := bytes.NewReader(jsonData)
					_, e := io.Copy(w, reader)

					if e != nil {
						log.Fatal("Error unmarshaling JSON:", e)
					}

					close(res)
					return
				}
			case <-ctx.Done():
				reader := strings.NewReader("Request failed!\n")
				_, err := io.Copy(w, reader)

				if err != nil {
					http.Error(w, "Error writing response", http.StatusInternalServerError)
					return
				}
				return
			}
		}

	})

	// Start server on PORT
	fmt.Println("HTTP server started at http://localhost:8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}

}
