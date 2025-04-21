package main

import "fmt"

type MessageCode int

const (
	NewLoginDetails MessageCode = iota
	LoginDetailsRequired
	IncorrectLogin
	AuthenticationError
	AuthenticationRequired
	LoginSuccessful
	Welcome
	APIKey
	RequestTimeout
	FailedMessageSend
	ConnectionError
	DatabaseError
)

type Response interface {
	GetMessage() string
}

type AuthResponse struct {
	Message string      `json:"message"`
	Code    MessageCode `json:"code"`
}

func (a AuthResponse) GetMessage() string {
	return fmt.Sprintf("Auth message %q", a.Message)
}

type ClientResponse struct {
	Err     *RequestError `json:"error"`
	Message string        `json:"message"`
	Code    MessageCode   `json:"code"`
}

func (c ClientResponse) GetMessage() string {
	return fmt.Sprintf("Message %q", c.Message)
}
