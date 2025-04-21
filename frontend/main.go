package main

import (
	"sync"

	"github.com/rivo/tview"
)

// Messages between parts of the app triggering different actions/events
type AppMessage struct {
	Code    MessageCode
	Message string
	Payload interface{}
}

type appState struct {
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

func NewAppState() *appState {
	return &appState{
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

	myAppState := NewAppState()

	app := tview.NewApplication()

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

func messageBroker(s *appState) {

	// App loop
	for {
		select {
		case m := <-s.UIBroadcast:
			for _, sub := range s.UISubscriptions {

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

		}

	}

}
