package main

import (
	"fmt"
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Cycle between pages of search, and chat with friends

type FriendsScreenPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Pages
	UIChannels
}

func (f *FriendsScreenPrimitive) End() {
	f.done <- struct{}{}
}

func (f *FriendsScreenPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func FriendsPages(s *appState) IOPrimitive {

	pages := tview.NewPages()
	pages.SetBorder(true)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	friendPages := FriendsScreenPrimitive{
		prim:       pages,
		UIChannels: uiCh,
	}

	// Front page
	list := tview.NewList().
		AddItem("Search", "Search for new friends", 's', func() {
			pages.SwitchToPage("Search")

			if s.loggedIn {

				friendPages.UIMessage <- &AppMessage{
					Code:    SearchUsers,
					Payload: nil,
					Message: "Type to search users",
				}
			}

		}).
		AddItem("Friends", "Chat with your friends", 'f', func() {
			pages.SwitchToPage("Friends")
		}).
		AddItem("Pending", "See pending friend requests", 'p', func() {
			pages.SwitchToPage("Pending")
		}).
		AddItem("Home", "Home (ctrl+c at any time)", 'x', func() {
			friendPages.UIMessage <- &AppMessage{
				Code:    Home,
				Payload: nil,
				Message: "Returned to home screen",
			}

		})
	list.SetBorder(true)

	// Direct focus to list
	pages.SetFocusFunc(func() {
		s.app.SetFocus(list)
	})

	// Search page
	search := SearchScreen(s)

	// Friends page
	var friends IOPrimitive
	friends = FriendsScreen(s)

	// Pending friends page
	var pending IOPrimitive
	pending = PendingScreen(s)

	// Configuring pages behavior
	pages.AddPage("List", list, true, true)
	pages.AddPage("Search", search.GetPrim(), true, false)
	pages.AddPage("Friends", friends.GetPrim(), true, false)
	pages.AddPage("Pending", pending.GetPrim(), true, false)

	pages.SetBorder(false)
	pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		name := event.Name()
		switch name {
		case "Home", "Esc":

			pages.SwitchToPage("List")
			s.app.SetFocus(list)
			friendPages.UIMessage <- &AppMessage{
				Code:    GameStart,
				Payload: nil,
				Message: "",
			}

			return nil
		}
		return event
	})

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(friendPages.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-friendPages.RecUIMess:

				switch m.Code {
				case OpenChat:
					// Get chat details
					var username string
					m.DecodePayload(&username)

					chatLog, ok := s.messages[username]

					if !ok {
						return
					}

					// Create new chat screen with content
					pages.RemovePage("Chat")

					c := ChatScreen(s, &chatLog)
					pages.AddAndSwitchToPage("Chat", c.GetPrim(), true)

				default:
					// Do nothing
				}

			case <-friendPages.done:
				break
			}
		}

	}()

	// Games page
	return &friendPages
}

/*
Scrollable view with usernames, and frames to select.
Only displays ten results, an closest matching.
Focusses on input to prompt search.
Type y to add or n to decline. Send a message for a request.

1) Type usernames and prompt backend for search results -- DONE
2) Display list of users in scrollable view
3) Displays username and some text as a bio
4) click to prompt user to add or not
5) If add, send friend request. Backen keeps track of friend requests etc
6) If other user accepts, the backend receives this and forwards on the successful add
7) If rejected, then the user will be notified of failure of friend request
*/

type SearchScreenPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Grid

	UIChannels
}

func (f *SearchScreenPrimitive) End() {
	f.done <- struct{}{}
}

func (f *SearchScreenPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func ResultBoxFac(n string, net chan *AppMessage) *tview.Frame {

	txt := tview.NewTextView()
	txt.SetText(fmt.Sprintf("%q\nDo you want to add friend?(y)", n))
	txt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y':
			// Send network message
			appMess := AppMessage{
				Code:    FriendRequest,
				Payload: nil,
				Message: "Friend request sent",
			}

			appMess.EncodePayload(&n)

			net <- &appMess
			return nil
		}
		return event
	})

	txt.SetBorderPadding(0, 0, 0, 0)

	frame := tview.NewFrame(
		txt,
	)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

func SearchScreen(s *appState) IOPrimitive {

	grid := tview.NewGrid().SetMinSize(7, 5)
	grid.SetBorder(true)

	resultsArr := []*tview.Frame{}

	hasFocus := 0
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {

		case tcell.KeyUp:

			if hasFocus-1 >= 0 {
				hasFocus -= 1
				s.app.SetFocus(resultsArr[hasFocus])
			}

			return nil

		case tcell.KeyDown:

			if hasFocus+1 < len(resultsArr) {
				hasFocus += 1
				s.app.SetFocus(resultsArr[hasFocus])
			}
			return nil

		}
		return event
	})

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	search := SearchScreenPrimitive{
		prim:       grid,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(search.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-search.RecUIMess:

				switch m.Code {
				case SearchUsersResults:
					// Clear results list
					for _, p := range resultsArr {
						grid.RemoveItem(p)
					}

					// DEcode results
					var results UsersSearch
					m.DecodePayload(&results)

					// Set header
					grid.SetTitle(fmt.Sprintf("Results: %d", len(results)))

					for i, n := range results {
						resultBox := ResultBoxFac(n, search.NetworkMessage)
						resultBox.SetFocusFunc(func() {
							hasFocus = i
						})
						grid.AddItem(resultBox, i, 0, 1, 1, 1, 1, false)
						resultsArr = append(resultsArr, resultBox)

					}

					hasFocus = 0
					s.app.SetFocus(resultsArr[0])
				default:
					// Do nothing
				}

			case <-search.done:
				break
			}
		}

	}()

	return &search
}

// List of all friends, and their active status
func FriendFac(n *Friend, UIBroadcast chan *AppMessage) *tview.Frame {

	var activeText string
	var borderColor tcell.Color

	if n.Active {
		activeText = "is active. Message? (y)"
		borderColor = tcell.ColorGreen
	} else {
		activeText = "is inactive. Message? (y)"
		borderColor = tcell.ColorDarkRed
	}

	txt := tview.NewTextView()
	txt.SetText(fmt.Sprintf("%v %v", n.Username, activeText))

	txt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y':
			// Send app message to
			appMess := AppMessage{
				Code:    OpenChat,
				Payload: nil,
			}

			appMess.EncodePayload(n.Username)

			UIBroadcast <- &appMess
			return nil
		}
		return event
	})

	frame := tview.NewFrame(
		txt,
	)

	frame.SetBorderPadding(0, 0, 0, 0).SetBorderColor(borderColor)
	frame.SetBorder(true)

	return frame
}

type FriendListPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Grid

	UIChannels
}

func (f *FriendListPrimitive) End() {
	f.done <- struct{}{}
}

func (f *FriendListPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func FriendsScreen(s *appState) IOPrimitive {

	grid := tview.NewGrid().SetMinSize(7, 5)
	grid.SetBorder(true)

	resultsArr := []*tview.Frame{}
	hasFocus := 0
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {

		case tcell.KeyUp:

			if hasFocus-1 >= 0 {
				hasFocus -= 1
				s.app.SetFocus(resultsArr[hasFocus])
			}

			return nil

		case tcell.KeyDown:

			if hasFocus+1 < len(resultsArr) {
				hasFocus += 1
				s.app.SetFocus(resultsArr[hasFocus])
			}
			return nil

		}
		return event
	})

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	list := FriendListPrimitive{
		prim:       grid,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(list.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-list.RecUIMess:

				switch m.Code {
				case AllContent:
					// Set header
					for _, p := range resultsArr {
						grid.RemoveItem(p)
					}
					for i, n := range s.friends {
						resultBox := FriendFac(&n, list.UIMessage)
						resultBox.SetFocusFunc(func() {
							hasFocus = i
						})
						resultsArr = append(resultsArr, resultBox)
						grid.AddItem(resultBox, i, 0, 1, 1, 1, 1, false)
					}

					hasFocus = 0
					s.app.SetFocus(resultsArr[0])
				case UpdateFriendContent:
					// Set header
					for _, p := range resultsArr {
						grid.RemoveItem(p)
					}
					for i, n := range s.friends {

						resultBox := FriendFac(&n, list.UIMessage)
						resultBox.SetFocusFunc(func() {
							hasFocus = i
						})
						resultsArr = append(resultsArr, resultBox)
						grid.AddItem(resultBox, i, 0, 1, 1, 1, 1, false)
					}

					hasFocus = 0
					s.app.SetFocus(resultsArr[0])

				default:
					// Do nothing
				}

			case <-list.done:
				break
			}
		}

	}()

	return &list
}

type PendingScreenPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Grid
	UIChannels
}

func (f *PendingScreenPrimitive) End() {
	f.done <- struct{}{}
}

func (f *PendingScreenPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func RequestBoxFac(n *FriendReqDetails, net chan *AppMessage) *tview.Frame {

	txt := tview.NewTextView()

	var displayTxt string
	var displayBorderCol tcell.Color
	if n.FromClient {
		displayTxt = fmt.Sprintf("%v\nFriend request pending", n.Username)
		displayBorderCol = tcell.ColorOrange
	} else {
		displayTxt = fmt.Sprintf("Friend request from %v: accept? (y/n)", n.Username)
		displayBorderCol = tcell.ColorBlue

	}
	txt.SetText(displayTxt)
	txt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'n':
			// Send message to backend with rejection
			aMess := AppMessage{
				Code:    FriendAccept,
				Payload: nil,
				Message: "Reject friend request",
			}

			b := FriendAcceptData{
				Accept:    false,
				RequestId: n.RequestId,
			}

			aMess.EncodePayload(&b)

			net <- &aMess

			return nil
		case 'y':
			// Send message to backend with acceptance
			aMess := AppMessage{
				Code:    FriendAccept,
				Payload: nil,
				Message: "Accept friend request",
			}

			b := FriendAcceptData{
				Accept:    true,
				RequestId: n.RequestId,
			}

			aMess.EncodePayload(&b)

			net <- &aMess
			return nil
		}
		return event
	})

	frame := tview.NewFrame(
		txt,
	)
	frame.SetBorderPadding(0, 0, 0, 0).SetBorderColor(displayBorderCol)
	frame.SetBorder(true)

	return frame
}

func PendingScreen(s *appState) IOPrimitive {

	grid := tview.NewGrid().SetMinSize(7, 5)
	grid.SetBorder(true)

	resultsArr := []*tview.Frame{}
	hasFocus := 0
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		switch event.Key() {

		case tcell.KeyUp:

			if hasFocus-1 >= 0 {
				hasFocus -= 1
				s.app.SetFocus(resultsArr[hasFocus])
			}

			return nil

		case tcell.KeyDown:

			if hasFocus+1 < len(resultsArr) {
				hasFocus += 1
				s.app.SetFocus(resultsArr[hasFocus])
			}
			return nil

		}
		return event
	})

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	search := PendingScreenPrimitive{
		prim:       grid,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(search.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-search.RecUIMess:

				switch m.Code {
				case UpdateFriendContent:
					hasFocus = 0
					// Set header
					for _, p := range resultsArr {
						grid.RemoveItem(p)
					}
					for i, n := range s.friendRequests {

						resultBox := RequestBoxFac(&n, s.networkBroadcast)
						resultBox.SetFocusFunc(func() {
							hasFocus = i
						})
						resultsArr = append(resultsArr, resultBox)
						grid.AddItem(resultBox, i, 0, 1, 1, 1, 1, false)
					}

					hasFocus = 0
					s.app.SetFocus(resultsArr[0])

				case AllContent:

					hasFocus = 0
					// Set header
					for _, p := range resultsArr {
						grid.RemoveItem(p)
					}
					for i, n := range s.friendRequests {

						resultBox := RequestBoxFac(&n, s.networkBroadcast)
						resultBox.SetFocusFunc(func() {
							hasFocus = i
						})
						resultsArr = append(resultsArr, resultBox)
						grid.AddItem(resultBox, i, 0, 1, 1, 1, 1, false)
					}

					hasFocus = 0
					s.app.SetFocus(resultsArr[0])
				default:
					// Do nothing
				}

			case <-search.done:
				break
			}
		}

	}()

	return &search

}

type ChatScreenPrimitive struct {
	// Reference to underlying primitive
	prim *tview.TextView
	UIChannels
}

func (f *ChatScreenPrimitive) End() {
	f.done <- struct{}{}
}

func (f *ChatScreenPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func ChatScreen(s *appState, chatLog *[]Message) IOPrimitive {
	txt := tview.NewTextView()
	txt.SetBorder(true)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	search := ChatScreenPrimitive{
		prim:       txt,
		UIChannels: uiCh,
	}

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(search.RecUIMess, UI)

	if err != nil {
		log.Fatal(err)
	}

	log := ""
	for _, c := range *chatLog {
		log += fmt.Sprintf("%v: %q\n", c.Date, c.Text)
	}
	txt.SetText(log)

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-search.RecUIMess:

				switch m.Code {
				// Wait for new text to appear

				default:
					//Do nothing

				}

			case <-search.done:
				break
			}
		}

	}()

	return &search

}
