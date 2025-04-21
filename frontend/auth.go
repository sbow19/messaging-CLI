package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"io"
)

type apiKey string

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

// Interface defines locations where content is displayed asynchronously, like from networ calls
type IOScreen interface {
	// Which component does the output of text go
	SetOutput(l io.Writer) error

	// Input element location
	SetInput(l io.Reader) error

	ReadInput() string

	Write(d []byte) error
}

type MainUserPrompt struct {
	input  *bufio.Scanner
	output io.Writer
}

func (u *MainUserPrompt) SetInput(s *bufio.Scanner) error {
	u.input = s
	return nil
}

func (u *MainUserPrompt) SetOutput(l io.Writer) error {
	u.output = l
	return nil
}

func (u *MainUserPrompt) ReadInput() string {
	// Blocks until user input provided
	u.input.Scan()
	return u.input.Text()
}

// Write content to screen
func (u *MainUserPrompt) Write(data []byte) error {
	u.output.Write(data)

	return nil
}

// Main user prompt area in home screen
func promptLoginDetails(screen IOScreen) (*LoginDetails, error) {
	screen.Write([]byte("Enter username\n"))
	username := screen.ReadInput()

	screen.Write([]byte("Enter password\n"))
	password := screen.ReadInput()

	login := &LoginDetails{
		Username: string(username),
		Password: string(password),
	}

	return login, nil

}
