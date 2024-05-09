package cli

import (
	"fmt"
	"slices"
	"strings"

	gui "github.com/grupawp/warships-gui/v2"
)

type GameUI struct {
	Controller *gui.GUI
	PBoard     *GameBoard
	OppBoard   *GameBoard
	EndText    *gui.Text
	TurnText   *gui.Text
	ErrorText  *gui.Text
	Timer      *gui.Text
}

func InitGameUI() *GameUI {
	ui := GameUI{
		Controller: gui.NewGUI(true),
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
		ui.OppBoard.Nick,
		ui.OppBoard.Board,
		ui.PBoard.Nick,
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

func (ui *GameUI) PlaceAndDrawDescriptions(pDesc string, oppDesc string) {
	pDescText := SplitLongText(ui.PBoard.xCoord, ui.PBoard.yCoord, 42, pDesc)
	oppDescText := SplitLongText(ui.OppBoard.xCoord, ui.OppBoard.yCoord, 42, oppDesc)
	drawables := slices.Concat(pDescText, oppDescText)
	for _, drawable := range drawables {
		ui.Controller.Draw(drawable)
	}
}

func SplitLongText(xCoord, yCoord, charsOnLine int, text string) []*gui.Text {
	words := strings.Split(text, " ")
	textBlobs := make([]*gui.Text, 0)
	currentLineChars := -1
	lineCoord := 23 + yCoord
	prevChar := 0
	for n, word := range words {
		wordLen := len(word)
		currentLineChars = currentLineChars + wordLen + 1
		if currentLineChars > charsOnLine && n != prevChar {
			textLine := strings.Join(words[prevChar:n], " ")
			textBlobs = append(textBlobs, gui.NewText(xCoord, lineCoord, textLine, nil))
			prevChar = n
			lineCoord++
			currentLineChars = -1
		}
		if wordLen > charsOnLine {
			splittedWord := splitWord(charsOnLine, word)
			for _, wordPart := range splittedWord {
				textBlobs = append(textBlobs, gui.NewText(xCoord, lineCoord, wordPart, nil))
				lineCoord++
			}
			prevChar = n + 1
			currentLineChars = -1
		}
	}
	textLine := strings.Join(words[prevChar:], " ")
	textBlobs = append(textBlobs, gui.NewText(xCoord, lineCoord, textLine, nil))
	return textBlobs
}

func splitWord(charsOnLine int, word string) []string {
	wordLen := len(word)
	splitGap := charsOnLine - 1
	splittedParts := make([]string, 0)
	for i := 0; i < wordLen; i += splitGap {
		if lastIndex := i + splitGap; lastIndex <= wordLen {
			splittedParts = append(splittedParts, word[i:lastIndex]+"-")
		} else {
			splittedParts = append(splittedParts, word[i:])
		}
	}
	return splittedParts
}
