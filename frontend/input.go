package main

import (
	"context"
	"log"

	"github.com/rivo/tview"
)

type FramePrimitive struct {
	// Reference to underlying primitive
	prim *tview.Frame

	UIChannels
}

func (f *FramePrimitive) End() {
	f.done <- struct{}{}
}

func (f *FramePrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func InputBar(s *appState) IOPrimitive {

	// Input part
	textarea := tview.NewTextArea().SetPlaceholder("Type...")
	textarea.SetBorder(true)

	// Input frame, prompt changes
	frame := tview.NewFrame(textarea).
		SetBorders(1, 0, 1, 0, 1, 1)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		NetworkMessage: s.networkBroadcast,
		UIMessage:      s.UIBroadcast,
		done:           make(chan struct{}),
	}

	input := FramePrimitive{
		prim:       frame,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(input.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	var (
		cancelPrompt context.CancelFunc
	)

	// Listen to UI broadcasts
	go func() {
		for {
			select {
			case m := <-input.RecUIMess:

				switch m.Code {
				// Prompt login details
				case LoginDetailsRequired:
					// Cancel any previous prompt
					if cancelPrompt != nil {
						cancelPrompt()
					}

					// Create a new context for this message
					var ctx context.Context
					ctx, cancelPrompt = context.WithCancel(context.Background())

					loginDetails := LoginDetails{
						Username: "",
						Password: "",
					}

					questions := Questions{
						&Question{
							q: "Please type your username",
							ref: func(input string) {
								loginDetails.Username = input
							},
						},
						&Question{
							q: "Please type your password",
							ref: func(input string) {
								loginDetails.Password = input
							},
						},
					}
					go PromptFlow(ctx, m.Code, &questions, m.Message, textarea, input.NetworkMessage, input.prim, &loginDetails)
				case SearchUsers, SearchUsersResults:
					if !s.loggedIn {
						continue
					}
					// Cancel any previous prompt
					if cancelPrompt != nil {
						cancelPrompt()
					}

					// Create a new context for this message
					var ctx context.Context
					ctx, cancelPrompt = context.WithCancel(context.Background())
					user := ""

					questions := Questions{
						&Question{
							q: "Please type a username",
							ref: func(input string) {
								user = input
							},
						},
					}
					go PromptFlow(ctx, m.Code, &questions, m.Message, textarea, input.NetworkMessage, input.prim, &user)
				default:
					/*Do Nothing*/
				}

			case <-input.done:
				if cancelPrompt != nil {
					cancelPrompt()
				}
				break
			}
		}

	}()
	return &input
}
