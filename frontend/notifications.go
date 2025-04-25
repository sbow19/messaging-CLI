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

func NotificationBoxFac() *tview.Frame {
	frame := tview.NewFrame(
		tview.NewFlex(),
	)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

func NotificationsBar(s *appState) IOPrimitive {

	// Indicator bar
	bar := tview.NewBox().SetBackgroundColor(tcell.Color101)

	// Friends bar --> Add friends to list. TODO: scrollable?
	// Possibly paginate when clicking up and down arrows
	// Dynamically render when new friends message
	grid := tview.NewGrid().SetGap(0, 0).SetSize(1, 1, 1, 1)
	grid.SetBorderPadding(0, 0, 0, 0)
	grid.AddItem(
		NotificationBoxFac(), 1, 1, 1, 1, 1, 1, false,
	).AddItem(
		NotificationBoxFac(), 2, 1, 1, 1, 0, 0, false,
	).AddItem(
		NotificationBoxFac(), 3, 1, 1, 1, 0, 0, false,
	).AddItem(
		NotificationBoxFac(), 4, 1, 1, 1, 0, 0, false,
	).AddItem(
		NotificationBoxFac(), 5, 1, 1, 1, 0, 0, false,
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
