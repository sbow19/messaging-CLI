package main

import (
	"fmt"
	"sync"
	"time"
)

type Users map[apiKey]*clientData

var loggedMu = sync.Mutex{}

// Return all users
func (l *Users) allUsers() []string {
	loggedMu.Lock()
	defer loggedMu.Unlock()

	clientList := []string{}
	for _, v := range *l {
		clientList = append(clientList, v.loginDetails.Username)
	}
	return clientList
}

type clientData struct {
	message      string
	accountMade  bool
	username     string
	welcomeSent  bool
	loginDetails LoginDetails
	loggedIn     bool
	apiKey       apiKey
	active       bool
	err          error
	mu           sync.Mutex
	rwmu         sync.RWMutex
}

func (c *clientData) Read() string {
	c.rwmu.RLock()
	defer c.rwmu.RUnlock()
	return fmt.Sprintf("The client's status... %q", c.message)
}

func (c *clientData) Leave() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active {
		c.active = false
		c.loggedIn = false
		c.message = fmt.Sprintf("Inactive since %q", time.Now())
	}
}

func (c *clientData) LoginClient(k apiKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.active {
		c.active = true
		c.loggedIn = true
		c.message = fmt.Sprintf("Active since %q", time.Now())
	}
}

func (c *clientData) WelcomeSent() bool {
	// TODO: shared locks for read messages might be better here
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.welcomeSent
}

func (c *clientData) SetNewLogin(l *LoginDetails, k apiKey) *RequestError {

	// Copy login details before setting on current USerMap
	clientCopy := clientData{
		mu:          sync.Mutex{},
		rwmu:        sync.RWMutex{},
		accountMade: c.accountMade,
		message:     c.message,
		loggedIn:    c.loggedIn,
		apiKey:      c.apiKey,
		active:      c.active,
		err:         c.err,
		loginDetails: LoginDetails{
			Username: l.Username,
			Password: l.Password,
		},
	}

	// Try db operation first
	err := dbConn.UpdateClient(&clientCopy)

	if err != nil {
		return &RequestError{
			Message: "Failed to set new login details",
			Code:    DatabaseError,
		}
	}

	// Updte user amp on success
	c.mu.Lock()
	defer c.mu.Unlock()

	// Set user to logged in an active on first time
	c.loggedIn = true
	c.active = true

	return nil
}

func generateNewUser(k apiKey) *clientData {

	return &clientData{
		message: fmt.Sprintf("Newly created on %q", time.Now()),
		apiKey:  k,
		loginDetails: LoginDetails{
			Username: "",
			Password: "",
		},
		accountMade: false,
		active:      false,
		loggedIn:    false,
		err:         nil,
		welcomeSent: false,
		mu:          sync.Mutex{},
		rwmu:        sync.RWMutex{},
	}

}

// Protect the map with a  mutex
var UserMap Users = Users{
	"123456": {
		message: "dummy message",
		apiKey:  "123456",

		loginDetails: LoginDetails{
			Username: "hello",
			Password: "password",
		},
		accountMade: true,
		active:      true,
		loggedIn:    true,
		err:         nil,
		mu:          sync.Mutex{},
		rwmu:        sync.RWMutex{},

		welcomeSent: true,
	},
}
