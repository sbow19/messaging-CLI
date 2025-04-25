package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Games list
type GameScreenPrimitive struct {
	// Reference to underlying primitive
	prim *tview.Flex
	UIChannels
}

func (f *GameScreenPrimitive) End() {
	f.done <- struct{}{}
}

func (f *GameScreenPrimitive) GetPrim() tview.Primitive {
	return f.prim
}

func NewGamesView(s *appState) IOPrimitive {

	gamePages := tview.NewPages()

	frontFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(
			gamePages,
			0,
			15,
			false,
		)

	frontFlex.SetBorderPadding(0, 0, 0, 0)

	uiCh := UIChannels{
		RecUIMess:      make(chan *AppMessage, 3),
		NetworkMessage: s.networkBroadcast,
		UIMessage:      s.UIBroadcast,
		done:           make(chan struct{}),
	}
	games := GameScreenPrimitive{
		prim:       frontFlex,
		UIChannels: uiCh,
	}

	// Games options
	list := tview.NewList().
		AddItem("Snake", "", 's', func() {
			gamePages.SwitchToPage("Snake")
			// Update message box
			games.UIMessage <- &AppMessage{
				Code:    GameStart,
				Payload: nil,
				Message: "Press space to start",
			}

		}).
		AddItem("Invaders from Space", "", 'b', func() {
			gamePages.SwitchToPage("Space")
		}).
		AddItem("Tic-Tac-Toe", "", 't', func() {
			gamePages.SwitchToPage("Tic")

		}).
		AddItem("Home", "Go to home screen", 'x', func() {
			games.UIMessage <- &AppMessage{
				Code:    Home,
				Payload: nil,
				Message: "Returned to home screen",
			}
		})
	list.SetBorder(true)

	// Direct focus to list
	frontFlex.SetFocusFunc(func() {
		s.app.SetFocus(list)
	})

	// Snake game view
	var snake *tview.Table

	snake = NewSnakeScreen(s)

	//
	gamePages.AddPage("List", list, true, true)
	gamePages.AddPage("Snake", snake, true, false)
	gamePages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		name := event.Name()
		switch name {
		case "Home", "Esc":
			gamePages.SwitchToPage("List")
			games.UIMessage <- &AppMessage{
				Code:    GameStart,
				Payload: nil,
				Message: "",
			}

			gamePages.RemovePage("Snake")
			snake = NewSnakeScreen(s)
			gamePages.AddPage("Snake", NewSnakeScreen(s), true, false)

			return nil
		}
		return event
	})

	// Register primitive with UI broadcast handler
	err := s.SubscribeChannel(games.RecUIMess, UI)

	go func() {
		for {
			select {
			case m := <-games.RecUIMess:
				switch m.Code {
				// Some error with connection
				case AttemptLogin:

				case ConnectionError:

				case LoginDetailsRequired:

				default:
					/*Do nothing*/
				}

			case <-games.done:
				return
			}
		}
	}()

	if err != nil {
		log.Fatal("Cannot subscribe games to Broadcast")
	}

	return &games

}

type Position struct {
	X int
	Y int
}

func updateSnake(
	ctx context.Context,
	cancel context.CancelFunc,
	table *tview.Table,
	snake *[]Position,
	dir *direction,
	start *bool,
	food *bool,
	foodPosition *Position,
	points *int,
	state *appState,
) {
	// Clear all cells (or only ones that changed)
	for row := 0; row < 21; row++ {
		for col := 0; col < 25; col++ {
			table.GetCell(row, col).SetText(fmt.Sprintf("%-4s", " ")).SetBackgroundColor(tcell.ColorBlack)
		}
	}

	eaten := false
	// Food has been eaten!
	if !*food {
		//Create new food position
		eaten = true
		for {
			x := rand.Intn(25)
			y := rand.Intn(21)

			conflict := false
			for _, pos := range *snake {
				if x == pos.X && y == pos.Y {
					conflict = true
					break
				}
			}

			if !conflict {
				foodPosition.X = x
				foodPosition.Y = y
				table.GetCell(y, x).SetBackgroundColor(tcell.ColorBeige)
				*food = true
				break
			}

		}

	} else {
		table.GetCell(foodPosition.Y, foodPosition.X).SetBackgroundColor(tcell.ColorBeige)

	}

	if *start {

		finalTile := (*snake)[len((*snake))-1]

		head := Position{
			X: finalTile.X,
			Y: finalTile.Y,
		}

		//Update Snake position
		switch *dir {
		case up:
			head.Y = head.Y - 1

			if head.Y < 0 {
				head.Y = 21
			}

		case down:
			head.Y = head.Y + 1
			if head.Y > 21 {
				head.Y = 0
			}

		case left:
			head.X = head.X - 1
			if head.X < 0 {
				head.X = 25
			}
		case right:
			head.X = head.X + 1
			if head.X > 25 {
				head.X = 0
			}

		}

		// Does head conflict
		conflict := false
		for _, pos := range *snake {
			if head.X == pos.X && head.Y == pos.Y {
				conflict = true
				break
			}
		}
		if conflict {
			state.UIBroadcast <- &AppMessage{
				Code:    GameStart,
				Payload: nil,
				Message: fmt.Sprintf("GAME OVER! Final points: %d", *points),
			}
			cancel()
			return
		}

		// Does head meet a food position?
		if head.X == foodPosition.X && head.Y == foodPosition.Y {
			*points = *points + 1

			table.GetCell(foodPosition.Y, foodPosition.X).SetBackgroundColor(tcell.ColorBlack)
			*food = false

			state.UIBroadcast <- &AppMessage{
				Code:    GameStart,
				Payload: nil,
				Message: fmt.Sprintf("Points: %d", *points),
			}
		}

		if eaten {
			*snake = append((*snake), head)
		} else {
			*snake = append((*snake)[1:], head)

		}

		// Draw snake
		for _, p := range *snake {
			table.GetCell(p.Y, p.X).SetTextColor(tcell.ColorGreen).SetBackgroundColor(tcell.ColorDarkGreen)
		}
	}

	select {
	case <-ctx.Done():
		// Context was cancelled, exit early
		return
	default:
		// Non-blocking: continue execution
	}

}

type direction int

const (
	up direction = iota
	down
	left
	right
)

func SnakeStart() *[]Position {
	snake := []Position{
		Position{
			X: 5,
			Y: 5,
		},
		Position{
			X: 5,
			Y: 6,
		},
		Position{
			X: 5,
			Y: 7,
		},
		Position{
			X: 5,
			Y: 8,
		},
	}

	return &snake
}

func NewSnakeScreen(state *appState) *tview.Table {

	table := tview.NewTable().SetFixed(1, 1)
	table.SetBorders(true).SetBordersColor(tcell.ColorBlack)

	for row := 0; row < 21; row++ {
		for col := 0; col < 25; col++ {
			cell := tview.NewTableCell(fmt.Sprintf("%-2s", " \n"))
			table.SetCell(row, col, cell)
		}
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Listen for game star broadcasts
	go func() {

		newSnake := SnakeStart()

		// Snake movement direction
		direction := down
		start := false

		// Points
		points := 0
		food := true
		foodPosition := Position{
			X: 10,
			Y: 10,
		}

		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

			key := event.Key()
			switch key {
			case tcell.KeyRune:
				switch event.Rune() {
				case ' ':
					start = !start
					if start {
						state.UIBroadcast <- &AppMessage{
							Code:    GameStart,
							Payload: nil,
							Message: fmt.Sprintf("Points: %d", points),
						}
					} else {
						state.UIBroadcast <- &AppMessage{
							Code:    GameStart,
							Payload: nil,
							Message: "Press space to start",
						}
					}
				}
				return nil
			case tcell.KeyEsc, tcell.KeyHome:
				ticker.Stop()
				return nil
			case tcell.KeyUp:
				if (direction == up) || (direction == down) {
					// Do nothing
				} else {
					direction = up
				}
			case tcell.KeyDown:
				if (direction == up) || (direction == down) {
					// Do nothing
				} else {
					direction = down
				}
			case tcell.KeyLeft:
				if (direction == left) || (direction == right) {
					// Do nothing
				} else {
					direction = left
				}
			case tcell.KeyRight:
				if (direction == left) || (direction == right) {
					// Do nothing
				} else {
					direction = right
				}
			}
			return event
		})

		table.SetBlurFunc(func() {
			start = false

			state.UIBroadcast <- &AppMessage{
				Code:    GameStart,
				Payload: nil,
				Message: "",
			}
		})

		// Constantly running until leave screen
	ticker:
		for range ticker.C {
			state.app.QueueUpdateDraw(func() {
				updateSnake(
					ctx,
					cancel,
					table,
					newSnake,
					&direction,
					&start,
					&food,
					&foodPosition,
					&points,
					state,
				)
			})

			select {
			case <-ctx.Done():
				break ticker
			default:
				//Do nothing
			}
		}

	}()

	return table

}
