package main

import (
	"battleship_client/api/client"
	"battleship_client/logic"
	"context"

	wGui "github.com/RostKoff/warships-gui/v2"
)

func main() {
	controller := wGui.NewGUI(true)
	boardCh := make(chan []string)
	settingsCh := make(chan client.GameSettings)
	settings := client.GameSettings{}
	board := make([]string, 0)
	abort := make(chan rune)
	for {
		ctx, canc := context.WithCancel(context.Background())
		var char rune
		go logic.DisplayGameSettings(controller, settingsCh, &settings)
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case settings = <-settingsCh:
				logic.DisplayPlacement(controller, boardCh, abort)
			}

		}(ctx)
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case board = <-boardCh:
				settings.Coords = board
				logic.StartGame(controller, settings, abort)
			}
		}(ctx)
		go func(ctx context.Context) {
			select {
			case <-ctx.Done():
				return
			case char = <-abort:
				canc()
				return
			}
		}(ctx)
		controller.Start(ctx, nil)
		if char != ' ' {
			break
		}
	}
}
