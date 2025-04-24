package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ReadAPIKey reads the file and returns the API key.
func ReadAPIKey(filename string) (string, error) {
	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		return "", err // Handle file open errors
	}
	defer file.Close()

	// Read the file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for a line that starts with "API_KEY="
		if strings.HasPrefix(line, "API_KEY=") {
			// Extract the key after "API_KEY="
			return strings.TrimPrefix(line, "API_KEY="), nil
		}
	}

	// Handle case where no API_KEY was found
	if err := scanner.Err(); err != nil {
		return "", err // Handle scanner error
	}

	return "", fmt.Errorf("API key not found in the file")
}

type Questions []*Question

type Question struct {
	q   string             // Question to display
	ref func(input string) // reference to property in struct
}

func PromptFlow(ctx context.Context, code MessageCode, order *Questions, m string, input *tview.TextArea, output chan *AppMessage, qArea *tview.Frame, content interface{}) error {

	//Question numbers
	i := 0
	next := make(chan struct{})
	defer func() {
		qArea.Clear()

		// Reset input behaviour
		input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			return event
		})
		close(next)
	}()
	for i < len(*order) {

		question := (*order)[i]
		qArea.AddText(question.q, true, tview.AlignCenter, tcell.ColorWhite)

		input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			// Send input to next stage
			if event.Key() == tcell.KeyEnter {

				//Assign input to data structure
				question.ref(input.GetText())

				qArea.Clear()
				input.SetText("", false)

				// Go to next question
				i++
				next <- struct{}{}
				return nil
			} else {
				return event
			}
		})

	trash:
		for {
			select {
			case <-next:
				break trash
				// Do nothing
			case <-ctx.Done():
				return nil
			}
		}

	}

	var aMess AppMessage
	switch code {
	case LoginDetailsRequired:
		aMess = AppMessage{
			Message: "Login details",
			Payload: nil,
			Code:    AttemptLogin,
		}

	case SearchUsers:
		aMess = AppMessage{
			Message: "Search users",
			Payload: nil,
			Code:    SearchUsers,
		}

	}

	aMess.EncodePayload(content)

	// Short circuit if contxt cancelled
	select {
	case <-ctx.Done():
		return nil
	default:
		// Do nothing
	}

	// Broadcast message to network part of app
	output <- &aMess

	return nil
}
