// User login management

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type apiKey string

type LoginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func authenticationCycle(ctx context.Context, res chan<- ClientResponse, r *http.Request) bool {
	keyRaw := r.Header.Get("Authorization")

	// User must send Auth header with API key
	k, err := checkAuthValid(keyRaw)
	if err != nil {
		res <- ClientResponse{
			Err:     err,
			Message: "An error occured",
		}
		return false
	}

	// Check if user has login details
	if haveLogin := userHaveLogin(k); !haveLogin {
		// Log user in with new details
		loginDetails, err := decodeLoginDetails(ctx, r)
		if err != nil {
			res <- ClientResponse{
				Err: &RequestError{
					Message: "New Login required",
					Code:    NewLoginRequired,
				},
				Message: "New Login Required",
			}
			return false
		}

		// Set new login details
		UserMap[k].SetNewLogin(loginDetails, k)
	}

	// Check for logged in user
	if loggedIn := checkUserLoggedIn(k); !loggedIn {
		// Log user in with new details
		loginDetails, err := decodeLoginDetails(ctx, r)
		if err != nil {
			res <- ClientResponse{
				Err: &RequestError{
					Message: "Login required",
					Code:    LoginRequired,
				},
				Message: "Login Required",
			}
			return false
		}

		// Attempt login
		if !loginUser(loginDetails, k) {
			res <- ClientResponse{
				Err: &RequestError{
					Message: "Login details incorrect",
					Code:    LoginDetailsIncorrect,
				},
				Message: "Login details incorrect",
			}
			return false
		}
	}

	res <- ClientResponse{
		Err: &RequestError{
			Message: "Login successful, upgrade to websocket",
			Code:    LoginSuccessful,
		},
		Message: string(k),
	}

	return true
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

	fmt.Println(k)

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

func decodeLoginDetails(ctx context.Context, r *http.Request) (*LoginDetails, *RequestError) {
	var loginDetails LoginDetails

	// Does user have login details
	err := json.NewDecoder(r.Body).Decode(&loginDetails)
	if err != nil {
		// Prompt for new user login
		e := &RequestError{
			Message: "New login details required",
			Code:    NewLoginRequired,
		}
		return nil, e
	}

	return &loginDetails, nil
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
