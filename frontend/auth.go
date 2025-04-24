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
