package main

import (
	"battleship_client/api/client"
	"battleship_client/logic"
	"context"

	wGui "github.com/RostKoff/warships-gui/v2"
)

func main() {
	gui := wGui.NewGUI(true)
	ctx := context.Background()
	ch := make(chan client.GameSettings)
	go logic.DisplayGameSettings(gui, ch)
	go func() {
		settings := <-ch
		logic.StartGame(gui, settings)
	}()
	gui.Start(ctx, nil)
}
