package main

import "fmt"

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
