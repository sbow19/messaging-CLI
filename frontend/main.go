package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/rivo/tview"
)

// Messages between parts of the app triggering different actions/events
type AppMessage struct {
	Code    MessageCode
	Message string
	Payload json.RawMessage
}

// Encode and decode payloads depending on the code type
func (a *AppMessage) EncodePayload(p interface{}) error {

	switch a.Code {
	case AttemptLogin:
		// P is LoginDetails type
		if result, ok := p.(*LoginDetails); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}

	}

	return nil

}

// Pass in expected type and unmarshal into that type
func (a *AppMessage) DecodePayload(target interface{}) error {

	switch a.Code {
	case AttemptLogin:
		// P is LoginDetails type
		if _, ok := target.(*LoginDetails); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}

	}

	return nil
}

type appState struct {
	app *tview.Application
	// Whether connection socket with backend active
	connected bool
	// Logged in message from the backend
	loggedIn bool

	// Channels between networking and UI portion

	// Message for networking goroutine
	networkBroadcast chan *AppMessage
	// Channels to receive network broadcast
	networkSubscriptions []chan *AppMessage

	// MEssage for UI goroutine
	UIBroadcast chan *AppMessage
	//Channels to receive UI message
	UISubscriptions []chan *AppMessage

	// Done
	done chan struct{}

	// Altering app state
	mu   sync.Mutex
	rwmu sync.RWMutex
}

func NewAppState(app *tview.Application) *appState {
	return &appState{
		app: app,

		connected:            false,
		networkBroadcast:     make(chan *AppMessage),
		networkSubscriptions: []chan *AppMessage{},

		UIBroadcast:     make(chan *AppMessage),
		UISubscriptions: []chan *AppMessage{},

		done: make(chan struct{}),
		mu:   sync.Mutex{},
		rwmu: sync.RWMutex{},
	}
}

func (m *appState) OpenConnection() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.connected {
		m.connected = true
	}

	return nil
}

func (m *appState) CloseConnection() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.connected {
		m.connected = false
	}

	return nil
}

func (m *appState) SetLoggedIn() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loggedIn {
		m.loggedIn = true
	}

	return nil
}

func (m *appState) SetLoggedOut() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.loggedIn {
		m.loggedIn = false
	}

	return nil
}

type BroadcastTypes int

const (
	UI BroadcastTypes = iota
	Network
)

func (m *appState) SubscribeChannel(c chan *AppMessage, t BroadcastTypes) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch t {
	case UI:
		// Listeners for messages intended for UI part
		m.UISubscriptions = append(m.UISubscriptions, c)
	case Network:
		// Listeners for messages intended for network part
		m.networkSubscriptions = append(m.networkSubscriptions, c)
	}

	return nil
}

func main() {
	app := tview.NewApplication()

	myAppState := NewAppState(app)
	go logger(myAppState)

	// Mnage intra-app messages
	go messageBroker(myAppState)
	// Set up networking
	go dialBackend(myAppState)
	// Set up UI. Receive channels. Gene
	flex := getUI(myAppState)

	// Start UI
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}

}

func logger(s *appState) {
	// Choose a known TTY path (you need to know this or find it dynamically)
	ttyPath := "/dev/pts/1" // Change this to your actual terminal device!

	// Open that terminal's device file
	tty, err := os.OpenFile(ttyPath, os.O_WRONLY, 0600)
	if err != nil {
		log.Fatalf("Failed to open TTY: %v", err)
	}
	defer tty.Close()

	// MultiWriter to both current terminal and the other one
	mw := io.MultiWriter(tty)
	log.SetOutput(mw)

	<-s.done

}

func messageBroker(s *appState) {

	// App loop

	for {
		select {
		case m := <-s.UIBroadcast:

			for i, sub := range s.UISubscriptions {

				log.Println(m, sub, i)
				// Broadcast message on subscribed UI element
				sub <- m
			}
		case m := <-s.networkBroadcast:
			for _, sub := range s.networkSubscriptions {

				// Broadcast message on subscribed network element
				sub <- m
			}

		case <-s.done:
			// Shut down app gracefully
			fmt.Print("\033[H\033[2J")
			os.Exit(0)
			return

		}

	}

}
