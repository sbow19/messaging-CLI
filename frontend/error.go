package main

import "fmt"

type RequestError struct {
	Message string      `json:"error_message"`
	Code    MessageCode `json:"code"`
}

// Implement error interface on request error
func (e RequestError) Error() string {
	return fmt.Sprintf("Error code %d; %q", e.Code, e.Message)
}

// Request error implements Response Interface
func (r RequestError) GetMessage() string {
	return fmt.Sprintf("Error code %d; %q", r.Code, r.Message)

}
