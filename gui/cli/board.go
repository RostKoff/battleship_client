package cli

import (
	"context"
	"fmt"
	"strconv"

	gui "github.com/RostKoff/warships-gui/v2"
)

type GameBoard struct {
	xCoord int
	yCoord int
	Nick   *gui.Text
	Desc   []*gui.Text
	Board  *gui.Board
	states [10][10]gui.State
}

func InitGameBoard(x int, y int, cfg *gui.BoardConfig) *GameBoard {
	b := GameBoard{}
	b.xCoord = x
	b.yCoord = y
	b.Board = gui.NewBoard(x, y, cfg)
	b.Nick = gui.NewText(x, y+22, "", nil)
	b.Board.SetStates(b.states)
	return &b
}

func (b *GameBoard) UpdateState(coords string, state gui.State) error {
	c, err := ConvertCoords(coords)
	if err != nil {
		return fmt.Errorf("failed to convert coords: %w", err)
	}
	b.states[c[0]][c[1]] = state
	b.Board.SetStates(b.states)
	return nil
}

func (b *GameBoard) UpdateStateWithDigitCoords(letterCoord int, numCoord int, state gui.State) error {
	if letterCoord < 0 || letterCoord > 9 {
		return fmt.Errorf("letter coord is out of bounce")
	}
	if numCoord < 0 || numCoord > 9 {
		return fmt.Errorf("number coord is out of bounce")
	}
	b.states[letterCoord][numCoord] = state
	b.Board.SetStates(b.states)
	return nil
}

// ! Function to consider
func (b *GameBoard) ListenForShot() string {
	coords := ""
	for {
		coords = b.Board.Listen(context.Background())
		c, _ := ConvertCoords(coords)
		state := b.states[c[0]][c[1]]
		if state == gui.Empty || state == "" {
			break
		}
	}
	return coords
}

func ConvertCoords(coords string) ([2]int, error) {
	values := [2]int{}
	letterCoord := int(coords[0]) - 65
	if letterCoord < 0 || letterCoord > 9 {
		return values, fmt.Errorf("letter coord is out of bounce")
	}
	numCoord, err := strconv.Atoi(coords[1:])
	if err != nil {
		return values, fmt.Errorf("failed to convert number coord to integer: %w", err)
	}
	numCoord--
	if numCoord < 0 || numCoord > 9 {
		return values, fmt.Errorf("number coord is out of bounce")
	}
	values[0] = letterCoord
	values[1] = numCoord
	return values, nil
}
