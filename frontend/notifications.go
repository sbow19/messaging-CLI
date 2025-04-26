package main

import (
	"fmt"
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

func BlankBox() *tview.Frame {
	txt := tview.NewTextView()
	frame := tview.NewFrame(
		txt,
	)
	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame

}

func NotificationBoxFac(m *Message, UIBroadcast chan *AppMessage) *tview.Frame {

	txt := tview.NewTextView()
	txt.SetText(fmt.Sprintf("%v: %v    sent: %v ", m.Sender, m.Text, m.Date))

	txt.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y':
			// Send app message to
			appMess := AppMessage{
				Code:    OpenChat,
				Payload: nil,
			}

			appMess.EncodePayload(m.Sender)

			UIBroadcast <- &appMess
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

// Track messages on side bar an trigger conversation pane
func NotificationsBar(s *appState) IOPrimitive {

	// Friends bar --> Add friends to list
	grid := tview.NewGrid().SetGap(0, 0).SetMinSize(5, 5)
	grid.SetBorderPadding(0, 0, 0, 0).SetBorder(true)

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
		log.Println(err)
		log.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		resultsArr = append(resultsArr, BlankBox())
	}

	// Listen to UI broadcasts
	go func() {

		for {
			select {
			case m := <-friendBar.RecUIMess:
				switch m.Code {

				case ReceiveMessage:
					var message Message

					err := m.DecodePayload(&message)

					if err != nil {
						return
					}

					if message.Sender == s.username {
						break
					}

					// Create msg notification box
					if len(resultsArr) == 5 {
						resultsArr = resultsArr[1:]

					}
					notifBox := NotificationBoxFac(&message, friendBar.UIMessage)
					resultsArr = append(resultsArr, notifBox)

					// Clear Grid and re add messages. 5 Recent notifications
					for _, p := range resultsArr {
						grid.RemoveItem(p)
					}
					for i, n := range resultsArr {
						n.SetFocusFunc(func() {
							hasFocus = i
						})
						grid.AddItem(n, i, 0, 1, 1, 1, 1, false)
					}

					hasFocus = 0
					s.app.SetFocus(resultsArr[0])

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
