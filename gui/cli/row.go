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

// func (r Row) Size() (int, int) {
// 	width := 0
// 	height := 0
// 	for _, button := range r.buttons {
// 		w, h := button.Size()
// 		width += w
// 		if height < h {
// 			height = h
// 		}
// 	}
// 	return width, height
// }

func (r Row) Size() (int, int) {
	width := 0
	height := 0
	len := len(r.buttons)
	if len == 0 {
		return width, height
	}
	fX, fY := r.buttons[0].Position()
	fW, fH := r.buttons[0].Size()

	width += fW
	height += fH

	if len < 2 {
		return width, height
	}
	lastX := fX + fW
	lastY := fY + fH
	for _, btn := range r.buttons[1:] {
		x, y := btn.Position()
		w, h := btn.Size()

		width += w + (x - lastX)

		if currH := y + h; lastY < currH {
			height = currH - fY
		}
		lastX = x + w
		lastY = y + h
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
