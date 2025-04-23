package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UIChannels struct {

	// Receive UI messages
	RecUIMess chan *AppMessage

	// Post messages to network part
	NetworkMessage chan *AppMessage

	// Post UI to network part
	UIMessage chan *AppMessage

	// End app
	done chan struct{}
}

type IOPrimitive interface {
	// End app
	End()

	//Get Primitve
	GetPrim() tview.Primitive
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
			0, 4, false).
		AddItem(friendsBar.GetPrim(), 0, 1, false)

	pageSlice := []IOPrimitive{
		inputBar,
		display,
		friendsBar,
	}
	i := 0

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		// Cycle through main boxes using tab key, globally available
		if event.Name() == "Tab" {

			if i = i + 1; i > 2 {
				i = 0
			}
			state.app.SetFocus(pageSlice[i].GetPrim())

			return nil
		}
		return event
	})

	return flex
}
