package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

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
	case SearchUsers:
		// P is LoginDetails type
		if result, ok := p.(*string); ok {

			jsonData, err := json.Marshal(*result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case SendMessage:
		// P is LoginDetails type
		if result, ok := p.(*Chat); ok {

			jsonData, err := json.Marshal(*result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case FriendRequest:
		// P is LoginDetails type
		if result, ok := p.(*string); ok {

			jsonData, err := json.Marshal(*result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case OpenChat:
		// P is LoginDetails type
		if result, ok := p.(string); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case SearchUsersResults:
		// P is LoginDetails type
		if result, ok := p.(*UsersSearch); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case AllContent:
		// P is LoginDetails type
		if result, ok := p.(*UserContent); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case FriendAccept:
		// P is LoginDetails type
		if result, ok := p.(*FriendAcceptData); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case ReceiveMessage:
		// P is LoginDetails type
		if result, ok := p.(*Message); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			a.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case NotifyLogin:
		// P is LoginDetails type
		if result, ok := p.(string); ok {

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
	case SearchUsers:
		// P is LoginDetails type
		if _, ok := target.(string); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case OpenChat:
		// P is LoginDetails type
		if _, ok := target.(*string); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case SendMessage:
		// P is LoginDetails type
		if _, ok := target.(*Chat); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case SearchUsersResults:
		// P is LoginDetails type
		if _, ok := target.(*UsersSearch); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case FriendRequestResult:
		// P is LoginDetails type
		if _, ok := target.(*string); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case FriendAcceptResult:
		// P is LoginDetails type
		if _, ok := target.(*string); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case AllContent:
		// P is LoginDetails type
		if _, ok := target.(*UserContent); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case FriendAccept:
		// P is LoginDetails type
		if _, ok := target.(*FriendAcceptData); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case ReceiveMessage:
		// P is LoginDetails type
		if _, ok := target.(*Message); ok {

			err := json.Unmarshal(a.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case NotifyLogin:
		// P is LoginDetails type
		if _, ok := target.(*string); ok {

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

	username string

	// Channels between networking and UI portion

	// Message for networking goroutine
	networkBroadcast chan *AppMessage
	// Channels to receive network broadcast
	networkSubscriptions []chan *AppMessage

	// MEssage for UI goroutine
	UIBroadcast chan *AppMessage
	//Channels to receive UI message
	UISubscriptions []chan *AppMessage

	// All user content
	friends        []Friend
	friendRequests []FriendReqDetails
	messages       Messages

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

		username: "",

		UIBroadcast:     make(chan *AppMessage),
		UISubscriptions: []chan *AppMessage{},

		done: make(chan struct{}),
		mu:   sync.Mutex{},
		rwmu: sync.RWMutex{},
	}
}

func (m *appState) AssignAllContent(u *UserContent) error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.friends = u.Friends
	m.friendRequests = u.FriendRequests
	m.messages = u.Messages
	return nil

}

func (m *appState) SetFriendActiveStatus(u string, active bool) error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	newSlice := m.friends[:0]
	for _, f := range m.friends {
		if f.Username == u {
			f.Active = active
		}
		newSlice = append(newSlice, f)
	}

	return nil
}

func (m *appState) AssignFriendshipContent(u *UserContent) error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.friends = u.Friends
	m.friendRequests = u.FriendRequests
	return nil

}

func (m *appState) AppendMessage(u *Message) error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	switch u.Receiver {
	case m.username:
		m.messages[u.Sender] = append(m.messages[u.Sender], *u)
	default:
		m.messages[u.Receiver] = append(m.messages[u.Receiver], *u)

	}

	return nil

}

func (m *appState) AddUserChat(u string) error {
	m.rwmu.Lock()
	defer m.rwmu.Unlock()

	m.messages[u] = []Message{}
	return nil

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

func (m *appState) SetUsername(u string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.username = u

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

func (m *appState) UnsubscribeChannel(c chan *AppMessage, t BroadcastTypes) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch t {
	case UI:
		filtered := m.UISubscriptions[:0] // reuse the original slice memory
		for _, v := range m.UISubscriptions {
			if v != c {
				filtered = append(filtered, v)
			}
		}
	case Network:
		// Listeners for messages intended for network part
		filtered := m.networkSubscriptions[:0] // reuse the original slice memory
		for _, v := range m.networkSubscriptions {
			if v != c {
				filtered = append(filtered, v)
			}
		}
	}

	return nil
}

func main() {
	app := tview.NewApplication()

	myAppState := NewAppState(app)
	go logger(myAppState)

	// Mnage intra-app messages
	go messageBroker(myAppState)
	// Set up networking -->
	// Connect ticker
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			// Blocks until it returns, then retry takes place
			dialBackend(myAppState)
		}
	}()
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
			fmt.Print("\033[H\033[2J")
			os.Exit(0)
			return

		}

	}

}
