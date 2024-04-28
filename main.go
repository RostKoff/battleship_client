package main

import (
	gui "battleship_client/gui/wp"
	"battleship_client/http/client"
	"context"
	"fmt"
	"time"

	warships_gui "github.com/grupawp/warships-gui/v2"
)

func main() {
	gameUi := gui.InitGameUI()
	settings := client.GameSettings{AgainstBot: true}
	statusRes := client.StatusResponse{}
	game, err := client.InitGame(settings)
	if err != nil {
		fmt.Printf("Failed to initialise game")
		return
	}

	for {
		statusRes, err = game.Status()
		if err != nil || statusRes.Status != "game_in_progress" {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	board, err := game.Board()
	if err != nil {
		fmt.Printf("Failed to get player board")
		return
	}
	for _, coord := range board {
		gameUi.PBoard.UpdateState(coord, warships_gui.Ship)
	}
	go func() {
		for {
			gameUi.Timer.SetText(fmt.Sprint(statusRes.Timer))
			if statusRes.Status == "ended" {
				break
			}
			if !statusRes.ShouldFire {
				gameUi.TurnText.SetText("")
				time.Sleep(time.Second)
			} else {
				if size := len(statusRes.OpponentShots); size > 0 {
					err = gameUi.HandleOppShot(board, statusRes.OpponentShots[size-1])
					if err != nil {
						fmt.Printf("Failed to draw opponent shot\n")
						return
					}
				}
				gameUi.TurnText.SetText("Your turn!")
				coord := gameUi.OppBoard.Listen()
				fireRes, err := game.Fire(coord)
				if err != nil {
					fmt.Printf("Failed to fire")
					return
				}
				err = gameUi.HandlePShot(fireRes, coord)
				if err != nil {
					fmt.Printf("Failed to handle player shot\n")
					return
				}
			}
			statusRes, err = game.Status()
			if err != nil {
				fmt.Printf("Cannot get game status!")
				return
			}
		}
		if statusRes.LastGameStatus == "lose" {
			gameUi.EndText.SetText("You lose!")
			return
		}
		gameUi.EndText.SetText("You won!")
	}()

	ctx := context.Background()
	gameUi.Controller.Start(ctx, nil)
}
