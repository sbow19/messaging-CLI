package main

import (
	"crypto/rand"
	"encoding/hex"
)

type LoginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Main user prompt area in home screen
func promptLoginDetails() (*LoginDetails, error) {
	// screen.Write([]byte("Enter username\n"))
	// username := screen.ReadInput()

	// screen.Write([]byte("Enter password\n"))
	// password := screen.ReadInput()

	// login := &LoginDetails{
	// 	Username: string(username),
	// 	Password: string(password),
	// }

	// return login, nil
	return nil, nil

}
