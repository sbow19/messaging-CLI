// User login management

package main

import (
	"encoding/base64"
	"net/http"
	"strings"
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

func getApiKey(r *http.Request) (apiKey, *RequestError) {
	keyRaw := r.Header.Get("Authorization")

	// User must send Auth header with API key
	k, err := checkAuthValid(keyRaw)
	if err != nil {
		return "", err
	}

	return k, nil
}

func doesUserExist(k apiKey) bool {

	_, ok := UserMap[k]

	// Add new user details to UserMap
	if !ok {
		clientData := generateNewUser(k)
		UserMap[k] = clientData
		return false
	}
	return true
}

func authenticationCycle(k apiKey, l *LoginDetails) (*AuthResponse, error) {

	// Check if there are login details
	if l.Password == "" || l.Username == "" {
		return &AuthResponse{
			Message: "Login details required",
			Code:    LoginDetailsRequired,
		}, nil
	}

	// Check if user has login details
	if haveLogin := userHaveLogin(k); !haveLogin {
		// Set new login details on client
		UserMap[k].SetNewLogin(l, k)
	}

	// Check for logged in user
	if loggedIn := checkUserLoggedIn(k); !loggedIn {
		// Attempt login
		if !loginUser(l, k) {
			return &AuthResponse{
				Message: "Login details incorrect",
				Code:    IncorrectLogin,
			}, nil
		}
	}

	return &AuthResponse{
		Message: "Login successful",
		Code:    LoginSuccessful,
	}, nil
}

func checkAuthValid(baseKey string) (apiKey, *RequestError) {
	// Check if it's Basic Auth
	if !strings.HasPrefix(baseKey, "Basic ") {
		// Auth error
		authError := &RequestError{
			Message: "Auth key incorrectly coded",
			Code:    AuthenticationError,
		}
		return "", authError
	}

	// Remove the "Basic " prefix
	keyValueEnc := baseKey[6:]

	// Decode the base64 part
	decoded, err := base64.StdEncoding.DecodeString(keyValueEnc)
	if err != nil {

		// Auth error
		authError := &RequestError{
			Message: "Auth key incorrectly coded",
			Code:    AuthenticationError,
		}
		return "", authError
	}

	// Split the decoded string into username and password
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		// Auth error
		authError := &RequestError{
			Message: "Auth key incorrectly coded",
			Code:    AuthenticationError,
		}
		return "", authError
	}

	var key apiKey = apiKey(parts[0])

	return key, nil

}

func userHaveLogin(k apiKey) bool {

	c := UserMap[k]

	// No username means detail required
	if c.loginDetails.Username == "" {
		return false
	}

	return true
}

func checkUserLoggedIn(k apiKey) bool {

	// Check user is logged in on the map
	c := UserMap[k]
	return c.loggedIn
}

func loginUser(l *LoginDetails, k apiKey) bool {
	// Get user details
	c := UserMap[k]

	if l.Username != c.loginDetails.Username {
		return false
	}

	if l.Password != c.loginDetails.Password {
		return false
	}

	// Change state of client data
	c.LoginClient(k)

	return true
}
