package main

import (
	"fmt"
	"sync"
	"time"
)

type ClientResponse struct {
	Err     *RequestError `json:"error"`
	Message string        `json:"message"`
}

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
	loginDetails LoginDetails
	loggedIn     bool
	apiKey       string
	active       bool
	err          error
	mu           sync.Mutex
}

func (c *clientData) Read() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fmt.Sprintf("The client's status... %q", c.message)
}

func (c *clientData) Leave() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active {
		c.active = false
		c.message = fmt.Sprintf("Inactive since %q", time.Now())
	}
}

func (c *clientData) LoginClient(k apiKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.active {
		c.active = true
		c.message = fmt.Sprintf("Active since %q", time.Now())
	}
}

func (c *clientData) SetNewLogin(l *LoginDetails, k apiKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.loginDetails.Username = l.Username
	c.loginDetails.Password = l.Password
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
		active:   true,
		loggedIn: true,
		err:      nil,
		mu:       sync.Mutex{},
	},
}
