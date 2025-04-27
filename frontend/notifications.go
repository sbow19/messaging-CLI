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

	return frame

}

func MessageNotificationBoxFac(m *Message, UIBroadcast chan *AppMessage) *tview.Frame {

	txt := tview.NewTextView().SetDynamicColors(true)
	txt.SetText(fmt.Sprintf("[blue::b]%v[white::-]: %v\nsent: %v\n Open chat?(y) ", m.Sender, m.Text, m.Date))

	frame := tview.NewFrame(
		txt,
	)
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y':
			// Send app message to
			appMess := AppMessage{
				Code:    OpenChat,
				Payload: nil,
			}

			appMess.EncodePayload(m.Sender)

			UIBroadcast <- &appMess
			return event
		}
		return event
	})

	frame.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {

		if event.Buttons() == tcell.Button1 {
			// Send app message to
			appMess := AppMessage{
				Code:    OpenChat,
				Payload: nil,
			}

			appMess.EncodePayload(m.Sender)

			UIBroadcast <- &appMess
			return action, event
		}
		return action, event
	})

	frame.SetBorderPadding(0, 0, 0, 0)
	frame.SetBorder(true)

	return frame
}

func OnlineNotificationBoxFac(user string, UIBroadcast chan *AppMessage) *tview.Frame {

	txt := tview.NewTextView().SetDynamicColors(true)
	txt.SetText(fmt.Sprintf("[green::b]%v[::-] is active.[white]\nOpen chat?(y) ", user))

	frame := tview.NewFrame(
		txt,
	)
	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'y':
			// Send app message to
			appMess := AppMessage{
				Code:    OpenChat,
				Payload: nil,
			}

			appMess.EncodePayload(user)

			UIBroadcast <- &appMess
			return event
		}
		return event
	})

	frame.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {

		if event.Buttons() == tcell.Button1 {
			// Send app message to
			appMess := AppMessage{
				Code:    OpenChat,
				Payload: nil,
			}

			appMess.EncodePayload(user)

			UIBroadcast <- &appMess
			return action, event
		}
		return action, event
	})

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
						break
					}

					if message.Sender == s.username {
						break
					}

					// Create msg notification box
					if len(resultsArr) == 5 {
						resultsArr = resultsArr[1:]

					}
					notifBox := MessageNotificationBoxFac(&message, friendBar.UIMessage)
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

				case NotifyLogin:
					var user string

					err := m.DecodePayload(&user)

					if err != nil {
						break
					}

					// Create msg notification box
					if len(resultsArr) == 5 {
						resultsArr = resultsArr[1:]

					}
					notifBox := OnlineNotificationBoxFac(user, friendBar.UIMessage)
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
