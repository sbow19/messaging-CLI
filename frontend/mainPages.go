package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

/*
Main content section
*/

type MessageBoxPrimitive struct {
	prim *tview.TextView
	UIChannels
}

func (f *MessageBoxPrimitive) End() {
	f.done <- struct{}{}
}

func (f *MessageBoxPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func NewMessageBox(s *appState) IOPrimitive {

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	messageBox := tview.NewTextView()

	box := MessageBoxPrimitive{
		prim:       messageBox,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(box.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-box.RecUIMess:
				switch m.Code {
				// Some error with connection
				case GameStart:
					messageBox.SetText("")
					messageBox.SetText(m.Message)
				case SearchUsers:
					messageBox.SetText("")
					messageBox.SetText(m.Message)
				case FriendRequestResult:
					var result string
					m.DecodePayload(&result)
					messageBox.SetText(result)
				case UpdateFriendContent:
					messageBox.SetText("")
					messageBox.SetText(m.Message)
				default:
					//Do nothing
				}

			case <-box.done:
				break
			}
		}

	}()

	return &box

}

func NewNetworkBox(s *appState) IOPrimitive {

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	messageBox := tview.NewTextView()

	box := MessageBoxPrimitive{
		prim:       messageBox,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(box.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-box.RecUIMess:
				switch m.Code {
				// Some error with connection
				case ConnectionError:
					messageBox.SetText("")
					messageBox.SetText(m.Message)
				case LoginDetailsRequired:
					messageBox.SetText("")
					messageBox.SetText(m.Message)
				case LoginSuccessful:
					messageBox.SetText("")
					messageBox.SetText(m.Message)
				default:
					//Do nothing
				}

			case <-box.done:
				break
			}
		}

	}()

	return &box

}

// About Page

func NewAboutPage() *tview.TextView {

	text := tview.NewTextView().SetText(
		`
		Hi, I'm Sam!

		Thank you for checking out this mini CLI project! I wanted to learn Go and thought this would
		be a great project to learn the language.

		You can connect with other people who use this CLI via teh friends section, and talk to them
		through the CLI!. You can also play some games, although they are single player only.

		Please reach out if you have any comments or want to collaborate on a project! My email is
		zctlsab@gmail.com
	`,
	).SetWordWrap(true)
	text.SetBorder(true).
		SetBorderPadding(1, 1, 1, 1).
		SetTitle("About Me!").
		SetTitleAlign(tview.AlignCenter)

	return text
}

type MainScreenPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Flex
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

	// About page
	text := NewAboutPage()

	// Games page
	games := NewGamesView(s)

	// Friends page - implements Search for new friends
	friends := FriendsPages(s)

	// Front page
	list := tview.NewList().
		AddItem("About", "Learn more about this project", 'h', func() {
			pages.SwitchToPage("About")
			s.app.SetFocus(text)

		}).
		AddItem("Friends", "Look for friends and chat", 'f', func() {
			pages.SwitchToPage("Friends")
			s.app.SetFocus(friends.GetPrim())
		}).
		AddItem("Games", "Play some terminal games", 'g', func() {
			pages.SwitchToPage("Games")
			s.app.SetFocus(games.GetPrim())

		}).
		AddItem("Exit", "End session (ctrl+c at any time)", 'x', func() {
			s.done <- struct{}{}
		})
	list.SetBorder(true)

	messageBox := NewMessageBox(s)
	networkBox := NewNetworkBox(s)

	frontFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			pages,
			0,
			15,
			false,
		).AddItem(
		networkBox.GetPrim(),
		1,
		1,
		false).AddItem(
		messageBox.GetPrim(),
		1,
		1,
		false,
	)

	// Direct focus to list
	frontFlex.SetFocusFunc(func() {
		s.app.SetFocus(list)
	})

	frontFlex.SetBorderPadding(0, 0, 0, 0)

	// Configuring pages behavior
	pages.AddPage("Home", list, true, true)
	pages.AddPage("Games", games.GetPrim(), true, false)
	pages.AddPage("About", text, true, false)
	pages.AddPage("Friends", friends.GetPrim(), true, false)

	pages.SetBorder(false)
	pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		name := event.Name()
		switch name {
		case "Home", "Esc":

			//Only Returns in top level parts of the app and games
			if !games.GetPrim().HasFocus() && !friends.GetPrim().HasFocus() {
				pages.SwitchToPage("Home")
			}
			return event
		}
		return event
	})

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	mainPages := MainScreenPrimitive{
		prim:       frontFlex,
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
			case m := <-mainPages.RecUIMess:

				switch m.Code {
				// Some error with connection
				case AttemptLogin:
				case ConnectionError:
				case LoginDetailsRequired:
				case Home:
					pages.SwitchToPage("Home")
				default:
					// Do nothing
				}

			case <-mainPages.done:
				break
			}
		}

	}()

	// Games page
	return &mainPages
}
