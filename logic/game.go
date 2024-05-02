package logic

import (
	"battleship_client/api/client"
	"battleship_client/gui/cli"
	"context"
	"fmt"
	"time"

	warships_gui "github.com/grupawp/warships-gui/v2"
)

func StartGame(gs client.GameSettings) {
	gameUi := cli.InitGameUI()
	statusRes := client.StatusResponse{}
	apiClient, err := client.InitGame(gs)
	if err != nil {
		fmt.Printf("Failed to initialise game\n")
		return
	}
	ctx := context.Background()

	for {
		statusRes, err = apiClient.Status()
		if err != nil {
			fmt.Print("Failed to get game status\n")
			return
		}
		if statusRes.Status != "game_in_progress" {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	board, err := apiClient.Board()
	if err != nil {
		fmt.Printf("Failed to get player board\n")
		return
	}
	for _, coord := range board {
		gameUi.PBoard.UpdateState(coord, warships_gui.Ship)
	}
	go func() {
		for {
			gameUi.SetTimer(statusRes.Timer)
			if statusRes.Status == "ended" {
				break
			}
			if !statusRes.ShouldFire {
				gameUi.TurnText.SetText("")
				time.Sleep(time.Second)
			} else {
				ctx2, cancel := context.WithCancel(ctx)
				cancel()
				go gameUi.DecreaseTimer(ctx2)
				gameUi.TurnText.SetText("Your turn!")
				if size := len(statusRes.OpponentShots); size > 0 {
					err = gameUi.HandleOppShot(board, statusRes.OpponentShots[size-1])
					if err != nil {
						fmt.Printf("Failed to draw opponent shot\n")
						return
					}
				}
				coord := gameUi.OppBoard.ListenForShot()
				fireRes, err := apiClient.Fire(coord)
				if err != nil {
					fmt.Printf("Failed to fire\n")
					return
				}
				err = gameUi.HandlePShot(fireRes, coord)
				if err != nil {
					fmt.Printf("Failed to handle player shot\n")
					return
				}
				cancel()
			}
			statusRes, err = apiClient.Status()
			if err != nil {
				fmt.Printf("Cannot get game status!\n")
				return
			}
		}
		if statusRes.LastGameStatus == "lose" {
			gameUi.EndText.SetText("You lose!\n")
			return
		}
		gameUi.EndText.SetText("You won!\n")
	}()

	gameUi.Controller.Start(ctx, nil)
}
