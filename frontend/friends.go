package main

import (
	"log"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type FriendsBarPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Grid

	UIChannels
}

func (f *FriendsBarPrimitive) End() {
	f.done <- struct{}{}
}

func (f *FriendsBarPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func FriendBoxFac() *tview.Frame {
	frame := tview.NewFrame(
		tview.NewFlex(),
	)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

func FriendsBar(s *appState) IOPrimitive {

	// Indicator bar
	bar := tview.NewBox().SetBackgroundColor(tcell.Color101)

	// Friends bar --> Add friends to list. TODO: scrollable?
	// Possibly paginate when clicking up and down arrows
	// Dynamically render when new friends message
	grid := tview.NewGrid().SetGap(0, 0).SetSize(1, 1, 1, 1)
	grid.SetBorderPadding(0, 0, 0, 0)
	grid.AddItem(
		FriendBoxFac(), 1, 1, 1, 1, 1, 1, false,
	).AddItem(
		FriendBoxFac(), 2, 1, 1, 1, 0, 0, false,
	).AddItem(
		FriendBoxFac(), 3, 1, 1, 1, 0, 0, false,
	).AddItem(
		FriendBoxFac(), 4, 1, 1, 1, 0, 0, false,
	).AddItem(
		FriendBoxFac(), 5, 1, 1, 1, 0, 0, false,
	).AddItem(
		bar, 6, 1, 1, 1, 0, 0, false,
	)
	grid.SetBorder(true)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		NetworkMessage: s.networkBroadcast,
		UIMessage:      s.UIBroadcast,
		done:           make(chan struct{}),
	}

	friendBar := FriendsBarPrimitive{
		prim:       grid,
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
			case m := <-friendBar.RecUIMess:
				switch m.Code {
				default:
					/*Do nothing*/
				}

			case <-friendBar.done:
				break
			}
		}

	}()
	return &friendBar
}

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

	// Pending friends page

	// Configuring pages behavior
	pages.AddPage("List", list, true, true)
	pages.AddPage("Search", search.GetPrim(), true, false)

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

func SearchScreen(s *appState) IOPrimitive {

	flex := tview.NewFlex()
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
					for _, n := range results {

						txt := tview.NewTextView()
						txt.SetText(n).SetBorder(true).SetBackgroundColor(tcell.ColorGold)
						flex.AddItem(
							txt, 0, 1, false,
						)
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

func FriendsScreen() {

}

func PendingScreen() {

}

func ChatScreen() {

}
