package cli

import (
	"context"
	"fmt"
	"slices"

	gui "github.com/RostKoff/warships-gui/v2"
)

type GameUI struct {
	Controller    *gui.GUI
	PBoard        *GameBoard
	OppBoard      *GameBoard
	EndText       *gui.Text
	TurnText      *gui.Text
	ErrorText     *gui.Text
	Timer         *gui.Text
	AbandonButton *gui.Button
}

func InitGameUI(controller *gui.GUI) *GameUI {
	ui := GameUI{
		Controller: controller,
		PBoard:     InitGameBoard(1, 5, nil),
		OppBoard:   InitGameBoard(50, 5, nil),
		EndText:    gui.NewText(50, 1, "", nil),
		TurnText:   gui.NewText(1, 1, "", nil),
		Timer:      gui.NewText(1, 3, "", nil),
		ErrorText:  gui.NewText(50, 3, "", nil),
	}

	ui.ErrorText.SetBgColor(gui.Red)
	ui.ErrorText.SetFgColor(gui.White)

	drawables := []gui.Drawable{
		ui.PBoard.Board,
		ui.PBoard.Nick,
		ui.PBoard.Desc,
		ui.OppBoard.Nick,
		ui.OppBoard.Board,
		ui.OppBoard.Desc,
		ui.EndText,
		ui.TurnText,
		ui.Timer,
		ui.ErrorText,
	}
	for _, drawable := range drawables {
		ui.Controller.Draw(drawable)
	}
	return &ui
}

func (ui *GameUI) HandleOppShots(pShips []string, oppShots []string) error {
	pShipsCopy := make([]string, len(pShips))
	copy(pShipsCopy, pShips)

	for _, shot := range oppShots {
		state := gui.Miss
		for n, ship := range pShipsCopy {
			if shot == ship {
				state = gui.Hit
				pShipsCopy = append(pShipsCopy[:n], pShipsCopy[n+1:]...)
				break
			}
		}
		err := ui.PBoard.UpdateState(shot, state)
		if err != nil {
			return fmt.Errorf("failed to update state: %w", err)
		}
	}
	return nil
}

func (ui *GameUI) HandlePShot(fireResponse string, coord string) error {
	switch fireResponse {
	case "hit":
		err := ui.OppBoard.UpdateState(coord, gui.Hit)
		if err != nil {
			return fmt.Errorf("failed to update board state: %w", err)
		}
	case "miss":
		err := ui.OppBoard.UpdateState(coord, gui.Miss)
		if err != nil {
			return fmt.Errorf("failed to update board state: %w", err)
		}
	case "sunk":
		c, err := ConvertCoords(coord)
		if err != nil {
			return fmt.Errorf("failed to convert coords: %w", err)
		}
		err = ui.OppBoard.UpdateState(coord, gui.Hit)
		if err != nil {
			return fmt.Errorf("failed to update board state: %w", err)
		}
		err = ui.handleSunk(c[0], c[1], nil)
		if err != nil {
			return fmt.Errorf("failed to handle sunk: %w", err)
		}
	default:
		return fmt.Errorf("unknown response")
	}
	return nil
}

// Recurrently fills the cells with the value "missed" around the sunken ship.
// Takes as arguments coordinates x and y of the cell with value "hit",
// and also a table with already visited hit cells.
func (ui *GameUI) handleSunk(x int, y int, checked *[][2]int) error {
	board := ui.OppBoard.states
	if board[x][y] == gui.Miss {
		return nil
	}
	if board[x][y] != gui.Hit {
		err := ui.OppBoard.UpdateStateWithDigitCoords(x, y, gui.Miss)
		if err != nil {
			return fmt.Errorf("failed to update board state: %w", err)
		}
		return nil
	}
	if checked == nil {
		s := make([][2]int, 0)
		checked = &s
	}
	*checked = append(*checked, [2]int{x, y})
	for i := x - 1; i <= x+1; i++ {
		if i < 0 || i > 9 {
			continue
		}
		for j := y - 1; j <= y+1; j++ {
			if j > 9 || j < 0 || slices.Contains(*checked, [2]int{i, j}) {
				continue
			}
			err := ui.handleSunk(i, j, checked)
			if err != nil {
				return fmt.Errorf("failed to handle sunk: %w", err)
			}
		}
	}
	return nil
}

func (ui *GameUI) DrawNicks(pNick string, oppNick string) {
	ui.PBoard.Nick.SetText(pNick)
	ui.OppBoard.Nick.SetText(oppNick)
}

func (ui *GameUI) DrawDescriptions(pDesc string, oppDesc string) {
	ui.PBoard.Desc.SetText(pDesc)
	ui.OppBoard.Nick.SetText(oppDesc)
}

// Listens for clicks on the opponent's board in a loop. Returns the coordinate of the clicked tile if it is empty.
func (ui *GameUI) ListenForShot(ctx context.Context) (string, error) {
	coords := ""
	// Loop until empty tile is clicked or context is done.
	for {
		coords = ui.OppBoard.Board.Listen(ctx)
		// Break if context is done.
		if coords == "" {
			break
		}
		// Check if empty tile is clicked.
		c, err := ConvertCoords(coords)
		if err != nil {
			return "", fmt.Errorf("failed to convert coords: %w", err)
		}
		state := ui.OppBoard.states[c[0]][c[1]]
		if state == "" {
			break
		}
	}
	return coords, nil
}
