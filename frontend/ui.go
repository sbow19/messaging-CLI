package main

import (
	"log"

	"github.com/rivo/tview"
)

type UIChannels struct {

	// Receive UI messages
	RecUIMess chan *AppMessage

	// Post messages to network part
	NetworkMessage chan *AppMessage

	// End app
	done chan struct{}
}

type IOPrimitive interface {
	// End app
	End()

	//Get Primitve
	GetPrim() tview.Primitive
}

/*
	Input bar
*/

// Receive broadcasts and change input style, input accordingly
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
		RecUIMess:      make(chan *AppMessage),
		NetworkMessage: s.networkBroadcast,
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

	// Listen to UI broadcasts
	go func() {
		for {
			select {
			case m := <-input.RecUIMess:

				switch m.Code {
				// Prompt login details
				case LoginDetailsRequired:

					loginDetails := &LoginDetails{
						Username: "",
						Password: "",
					}

					questions := Questions{
						&Question{
							q: "Please type username",
							ref: func(input string) {
								loginDetails.Username = input
							},
						},
						&Question{
							q: "Please type password",
							ref: func(input string) {
								loginDetails.Password = input
							},
						},
					}
					PromptFlow(&questions, m.Message, textarea, input.prim)

					if err != nil {
						// TODO: implement some prompt flow error
						break
					}

					// Broadcast message to network part of app
					input.NetworkMessage <- &AppMessage{
						Message: "Login details",
						Payload: loginDetails,
						Code:    AttemptLogin,
					}

				}
			case <-input.done:
				break
			}
		}

	}()
	return &input
}

/*
Main content section
*/
func FriendsScreen() *tview.Pages {
	pages := tview.NewPages()
	return pages
}

type MainScreenPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Pages

	UIChannels
}

func (f *MainScreenPrimitive) End() {
	f.done <- struct{}{}
}

func (f *MainScreenPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func MainScreenPages(s *appState) IOPrimitive {

	pages := tview.NewPages()

	// Front page
	list := tview.NewList().
		AddItem("About", "Learn more about this project", 'h', nil).
		AddItem("Friends", "Learn more about this project", 'f', nil).
		AddItem("Games", "Learn more about this project", 'g', nil).
		AddItem("Exit", "Learn more about this project", 'x', nil)

	pages.AddPage("Home", list, true, true)

	// About page
	text := tview.NewTextView()
	text.SetBorder(true)
	pages.AddPage("About", text, true, false)

	// Friends page - implements Search for new friends
	friends := FriendsScreen()
	pages.AddPage("Friends", friends, true, false)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage),
		NetworkMessage: make(chan *AppMessage),
		done:           make(chan struct{}),
	}

	mainPages := MainScreenPrimitive{
		prim:       pages,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(mainPages.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case <-mainPages.RecUIMess:

			case <-mainPages.done:
				break
			}
		}

	}()

	// Games page
	return &mainPages
}

/*
	Friends side bar
*/

type FriendsBarPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Flex

	UIChannels
}

func (f *FriendsBarPrimitive) End() {
	f.done <- struct{}{}
}

func (f *FriendsBarPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func FriendsBar(s *appState) IOPrimitive {

	// Indicator bar
	bar := tview.NewBox()

	// Friends bar --> Add friends to list. TODO: scrollable?
	grid := tview.NewGrid().
		SetRows(1).
		SetColumns(1)

	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(grid, 1, 5, false).
		AddItem(bar, 1, 1, false)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage),
		NetworkMessage: make(chan *AppMessage),
		done:           make(chan struct{}),
	}

	friendBar := FriendsBarPrimitive{
		prim:       flex,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(friendBar.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case <-friendBar.RecUIMess:

			case <-friendBar.done:
				break
			}
		}

	}()
	return &friendBar
}

// TODO: add channels to receive input from network calls
func getUI(state *appState) *tview.Flex {

	// Input and user prompt
	inputBar := InputBar(state)

	// Friends bar
	friendsBar := FriendsBar(state)

	// Main Display
	display := MainScreenPages(state)

	// App layout
	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(display.GetPrim(), 0, 5, false).
			AddItem(inputBar.GetPrim(), 0, 1, true),
			0, 4, true).
		AddItem(friendsBar.GetPrim(), 0, 1, false)

	return flex
}
