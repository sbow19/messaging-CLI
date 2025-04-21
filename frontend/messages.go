package main

import (
	"fmt"
	"io"
)

type Response interface {
	GetMessage() string
	GetCode() MessageCode
}

type MessageCode int

const (
	NewLoginDetails MessageCode = iota
	LoginDetailsRequired
	IncorrectLogin
	AuthenticationError
	AuthenticationRequired
	AttemptLogin
	LoginSuccessful
	Welcome
	APIKey
	RequestTimeout
	FailedMessageSend
	ConnectionError
	DatabaseError
)

type AuthResponse struct {
	Message string      `json:"message"`
	Code    MessageCode `json:"code"`
}

func (a AuthResponse) GetMessage() string {
	return fmt.Sprintf("Auth message %q", a.Message)
}
func (a AuthResponse) GetCode() MessageCode {
	return a.Code
}

type ClientResponse struct {
	Err     *RequestError `json:"error"`
	Message string        `json:"message"`
	Code    MessageCode   `json:"code"`
}

func (c ClientResponse) GetMessage() string {
	return fmt.Sprintf("Message %q", c.Message)
}
func (c ClientResponse) GetCode() MessageCode {
	return c.Code
}

// Client messages have a different type and interface
type ClientMessage struct {
	Payload io.Reader
	Code    ClientCode
}

type ClientCode int

const (
	Login ClientCode = iota
)
