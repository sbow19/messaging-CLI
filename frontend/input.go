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
	textarea := tview.NewTextArea().SetPlaceholder("Type...").SetSize(1, 100)
	textarea.SetBorder(true)

	// Input frame, prompt changes
	frame := tview.NewFrame(textarea).
		SetBorders(2, 0, 1, 0, 1, 1)

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

		//User to chat ith
		usr := ""

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
								s.SetUsername(input)
							},
						},
						&Question{
							q: "Please type your password",
							ref: func(input string) {
								loginDetails.Password = input
							},
						},
					}
					go PromptFlow(ctx, m.Code, &questions, m.Message, textarea, input.NetworkMessage, s.UIBroadcast, input.prim, &loginDetails)
				case SearchUsers, SearchUsersResults:
					if !s.loggedIn {
						break
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
					go PromptFlow(ctx, m.Code, &questions, m.Message, textarea, input.NetworkMessage, s.UIBroadcast, input.prim, &user)
				case GameStart:
					// Cancel any previous prompt
					if cancelPrompt != nil {
						cancelPrompt()
					}

					input.prim.Clear()
					textarea.SetText("", false)

				case OpenChat:
					// Cancel any previous prompt
					if cancelPrompt != nil {
						cancelPrompt()
					}

					textarea.SetText("", false)

					// Create a new context for this message
					var ctx context.Context
					ctx, cancelPrompt = context.WithCancel(context.Background())

					//Message object
					m.DecodePayload(&usr)
					chat := Chat{
						Text:     "",
						Receiver: usr,
						Sender:   s.username,
					}

					questions := Questions{
						&Question{
							q: "Type to chat",
							ref: func(input string) {
								chat.Text = input
							},
						},
					}
					go PromptFlow(ctx, SendMessage, &questions, m.Message, textarea, input.NetworkMessage, s.UIBroadcast, input.prim, &chat)

				case SendMessage:
					// Cancel any previous prompt
					if cancelPrompt != nil {
						cancelPrompt()
					}

					textarea.SetText("", false)

					// Create a new context for this message
					var ctx context.Context
					ctx, cancelPrompt = context.WithCancel(context.Background())

					//Message object
					chat := Chat{
						Text:     "",
						Receiver: usr,
						Sender:   s.username,
					}

					questions := Questions{
						&Question{
							q: "Type to chat",
							ref: func(input string) {
								chat.Text = input
							},
						},
					}
					go PromptFlow(ctx, SendMessage, &questions, m.Message, textarea, input.NetworkMessage, s.UIBroadcast, input.prim, &chat)

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
