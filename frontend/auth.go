package main

import (
	"crypto/rand"
	"encoding/hex"
)

type apiKey string

type LoginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

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
)

func generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
