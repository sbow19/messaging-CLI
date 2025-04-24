package main

import (
	"encoding/json"
	"fmt"
)

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
	Home
	GameStart
	SearchUsers
	SearchUsersResults
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
	Err     *RequestError   `json:"error"`
	Message string          `json:"message"`
	Code    MessageCode     `json:"code"`
	Payload json.RawMessage `json:"payload"`
}

func (c ClientResponse) GetMessage() string {
	return fmt.Sprintf("Message %q", c.Message)
}

// Encode and decode payloads depending on the code type
func (m *ClientResponse) EncodePayload(p interface{}) error {

	switch m.Code {
	case SearchUsersResults:
		// P is LoginDetails type
		if result, ok := p.(*UsersSearch); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			m.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	}

	return nil

}

// Pass in expected type and unmarshal into that type
func (m *ClientResponse) DecodePayload(target interface{}) error {

	switch m.Code {
	case SearchUsersResults:
		// P is LoginDetails type
		if _, ok := target.(*UsersSearch); ok {

			err := json.Unmarshal(m.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	}

	return nil
}

// Client messages have a different type and interface
type ClientMessage struct {
	Payload json.RawMessage `json:"payload"`
	Code    MessageCode     `json:"code"`
}

// Encode and decode payloads depending on the code type
func (m *ClientMessage) EncodePayload(p interface{}) error {

	switch m.Code {
	case AttemptLogin:
		// P is LoginDetails type
		if result, ok := p.(*LoginDetails); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			m.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}
	case SearchUsers:
		// P is LoginDetails type
		if result, ok := p.(string); ok {

			jsonData, err := json.Marshal(result)

			if err != nil {
				return err
			}

			m.Payload = jsonData

		} else {
			return fmt.Errorf("incorrect details")
		}

	}

	return nil

}

// Pass in expected type and unmarshal into that type
func (m *ClientMessage) DecodePayload(target interface{}) error {

	switch m.Code {
	case AttemptLogin:
		// P is LoginDetails type
		if _, ok := target.(*LoginDetails); ok {

			err := json.Unmarshal(m.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}
	case SearchUsers:
		// P is LoginDetails type
		if _, ok := target.(*string); ok {

			err := json.Unmarshal(m.Payload, target)

			if err != nil {
				return err
			}

		} else {
			return fmt.Errorf("incorrect details")
		}

	}

	return nil
}
