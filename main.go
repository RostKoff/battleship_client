package main

import (
	"battleship_client/http/client"
	"context"
	"time"

	gui "github.com/grupawp/warships-gui/v2"
)

func main() {

	ctx := context.Background()
	go func() {
		settings := client.GameSettings{AgainstBot: true}
		game, err := client.InitGame(settings)
		if err != nil {
			panic(err)
		}
		statusRes := client.StatusResponse{}
		for {
			statusRes, err = game.Status()
			if err != nil {
				continue
			} else if statusRes.Status == "game_in_progress" {
				break
			}
			time.Sleep(time.Second)
		}
		// descs, err := game.PlayerDescriptions()
		if err != nil {
			panic(err)
		}
		board, err := game.Board()
		if err != nil {
			panic(err)
		}
		pboard.Nick.SetText(statusRes.Nick)
		oppBoard.Nick.SetText(statusRes.Opponent)
		for _, coord := range board {
			err := pboard.UpdateState(coord, gui.Ship)
			if err != nil {
				panic(err)
			}
		}
	}()
	ui.Start(ctx, nil)

}
