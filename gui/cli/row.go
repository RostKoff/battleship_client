package cli

import (
	wGui "github.com/RostKoff/warships-gui/v2"
)

type Row struct {
	buttons []*wGui.Button
}

func NewRow(buttons []*wGui.Button) Row {
	return Row{buttons: buttons}
}

func (r Row) Position() (int, int) {
	if len(r.buttons) == 0 {
		return 0, 0
	}
	button := r.buttons[0]
	return button.Position()
}

func (r Row) Size() (int, int) {
	width := 0
	height := 0
	for _, button := range r.buttons {
		w, h := button.Size()
		width += w
		if height < h {
			height = h
		}
	}
	return width, height
}

func (r *Row) SetBgColor(color wGui.Color) {
	for _, button := range r.buttons {
		button.SetBgColor(color)
	}
}

func (r *Row) SetFgColor(color wGui.Color) {
	for _, button := range r.buttons {
		button.SetFgColor(color)
	}
}

func (r *Row) GetButtons() []*wGui.Button {
	return r.buttons
}
