package main

import "fmt"

type ErrorState int

// Error codes
const (
	LoginSuccessful ErrorState = iota
	LoginRequired
	NewLoginRequired
	LoginDetailsIncorrect
	AuthenticationRequired
	AuthenticationError
	RequestTimeout
)

type RequestError struct {
	Message string     `json:"error_message"`
	Code    ErrorState `json:"error_code"`
}

// Implement error interface on request error
func (e RequestError) Error() string {
	return fmt.Sprintf("Error code %d; %q", e.Code, e.Message)
}
