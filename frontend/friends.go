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
				case UpdateFriendContent:
					pages.RemovePage("Friends").RemovePage("Pending")
					friends = FriendsScreen(s)
					pending = PendingScreen(s)
					pages.AddPage("Friends", friends.GetPrim(), true, false)
					pages.AddPage("Pending", pending.GetPrim(), true, false)

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
	prim *tview.Flex

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
	txt.SetText(fmt.Sprintf("%q\n\nDo you want to add friend?(y)", n)).SetBorder(true)

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

	frame := tview.NewFrame(
		txt,
	)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

func ResultNoFac(no int) *tview.Frame {
	frame := tview.NewFrame(
		tview.NewTextView().SetText(fmt.Sprintf("Results: %d", no)),
	)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

func SearchScreen(s *appState) IOPrimitive {

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	search := SearchScreenPrimitive{
		prim:       flex,
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
					var results UsersSearch
					m.DecodePayload(&results)

					// Set header
					header := ResultNoFac(len(results))
					flex.AddItem(
						header, 0, 1, false,
					)

					for _, n := range results {
						resultBox := ResultBoxFac(n, search.NetworkMessage)
						flex.AddItem(resultBox, 0, 1, false)
					}
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
func FriendFac(n *Friend, net chan *AppMessage) *tview.Frame {

	var activeText string
	var borderColor tcell.Color

	if n.Active {
		activeText = "is active"
		borderColor = tcell.ColorGreen
	} else {
		activeText = "is inactive"
		borderColor = tcell.ColorDarkRed
	}

	txt := tview.NewTextView()
	txt.SetText(fmt.Sprintf("%v %v", n.Username, activeText)).
		SetBorder(true).
		SetBorderColor(borderColor)

	txt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// TODO: go to message screen
		return event
	})

	frame := tview.NewFrame(
		txt,
	)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

type FriendListPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Flex

	UIChannels
}

func (f *FriendListPrimitive) End() {
	f.done <- struct{}{}
}

func (f *FriendListPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func FriendsScreen(s *appState) IOPrimitive {

	var flex *tview.Flex
	flex = tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	list := FriendListPrimitive{
		prim:       flex,
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
					for _, n := range s.friends {
						resultBox := FriendFac(&n, list.UIMessage)
						flex.AddItem(resultBox, 0, 1, false)
					}

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
	prim *tview.Flex
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
	txt.SetText(displayTxt).SetBorder(true).SetBackgroundColor(displayBorderCol)
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
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

func PendingScreen(s *appState) IOPrimitive {

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		UIMessage:      s.UIBroadcast,
		NetworkMessage: s.networkBroadcast,
		done:           make(chan struct{}),
	}

	search := PendingScreenPrimitive{
		prim:       flex,
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
				case AllContent:

					// Set header
					for _, n := range s.friendRequests {
						resultBox := RequestBoxFac(&n, s.networkBroadcast)
						flex.AddItem(resultBox, 0, 1, false)
					}
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

func ChatScreen() {

}
