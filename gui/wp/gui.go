package wp

import (
	gui "github.com/grupawp/warships-gui/v2"
)

type GameUI struct {
	Controller *gui.GUI
	PBoard     *GameBoard
	OppBoard   *GameBoard
	EndText    *gui.Text
}

func InitGameUI() *GameUI {
	ui := GameUI{
		Controller: gui.NewGUI(false),
		PBoard:     InitGameBoard(1, 3, nil),
		OppBoard:   InitGameBoard(50, 3, nil),
		EndText:    gui.NewText(1, 1, "", nil),
	}

	drawables := []gui.Drawable{
		ui.PBoard.Board,
		ui.PBoard.Desc,
		ui.PBoard.Nick,
		ui.OppBoard.Board,
		ui.OppBoard.Desc,
		ui.PBoard.Nick,
		ui.EndText,
	}
	for _, drawable := range drawables {
		ui.Controller.Draw(drawable)
	}
	return &ui
}
