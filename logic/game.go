package logic

import (
	"battleship_client/api/client"
	"battleship_client/gui/cli"
	"context"
	"fmt"
	"time"

	warships_gui "github.com/grupawp/warships-gui/v2"
)

func StartGame(gs client.GameSettings) error {
	gameUi := cli.InitGameUI()
	statusRes := client.StatusResponse{}
	apiClient, err := client.InitGame(gs)
	errMsgChan := make(chan string)
	if err != nil {
		return fmt.Errorf("failed to initialise the game")
	}
	for errCount := 0; ; {
		statusRes, err = apiClient.Status()
		if err != nil {
			if errCount > 4 {
				return fmt.Errorf("failed to get game status")
			}
			gameUi.Controller.Log("Status Error %d: %s", errCount+1, err.Error())
			errCount++
			continue
		}
		errCount = 0
		if statusRes.Status != "game_in_progress" {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	board, err := apiClient.Board()
	if err != nil {
		return fmt.Errorf("failed to get player board")
	}
	// Fill the board with ships
	for _, coord := range board {
		gameUi.PBoard.UpdateState(coord, warships_gui.Ship)
	}
	gameUi.DrawNicks(statusRes.Nick, statusRes.Opponent)

	descs, err := apiClient.PlayerDescriptions()

	gameUi.PlaceAndDrawDescriptions(descs.PlayerDescription, descs.OpponentDescription)

	ctx := context.Background()
	// Context to cancel additional goroutine after game is finished.
	mainEnd, cancel := context.WithCancel(context.Background())
	// Goroutine that is responsible for handling information according to game status got from API.
	go func() {
		defer cancel()
		oppShotCount := 0
		for {
			statusRes, err = apiClient.Status()
			if err != nil {
				gameUi.Controller.Log(fmt.Sprintf("Status error: %s", err.Error()))
				errMsgChan <- "Failed to get game status"
				time.Sleep(time.Second)
				continue
			}
			if size := len(statusRes.OpponentShots); oppShotCount != size {
				err = gameUi.HandleOppShots(board, statusRes.OpponentShots)
				if err != nil {
					gameUi.Controller.Log("Handle opponent shots error: %s", err.Error())
					return
				}
				oppShotCount = size
			}
			if statusRes.Status == "ended" {
				break
			}
			if !statusRes.ShouldFire {
				gameUi.TurnText.SetText("Opponent Turn")
				gameUi.Timer.SetText("Time: -")
			} else {
				gameUi.Timer.SetText(fmt.Sprintf("Time: %d", statusRes.Timer))
				gameUi.TurnText.SetText("Your turn!")
			}
			time.Sleep(time.Second)
		}
		if statusRes.LastGameStatus == "lose" {
			gameUi.EndText.SetText("You lose!\n")
			return
		}
		gameUi.EndText.SetText("You won!\n")
	}()
	// Goroutine responsible for getting player input and handling fire mechanic.
	go func(ctx context.Context) {
	mainLoop:
		for {
			select {
			case <-ctx.Done():
				break mainLoop
			default:
				coord := gameUi.OppBoard.ListenForShot()
				fireRes, err := apiClient.Fire(coord)
				if err != nil {
					errMsgChan <- "Failed to fire"
					gameUi.Controller.Log(fmt.Sprintf("Fire error: %s", err.Error()))
					continue
				}
				err = gameUi.HandlePShot(fireRes, coord)
				if err != nil {
					gameUi.ErrorText.SetText("Failed to handle player shot\n")
					gameUi.Controller.Log(fmt.Sprintf("Player shot error: %s", err.Error()))
					continue
				}
			}
		}
	}(mainEnd)

	go func(ctx context.Context) {
		errTimer := time.NewTimer(time.Second * 10)
		errTimer.Stop()
	mainLoop:
		for {
			select {
			case <-ctx.Done():
				break mainLoop
			case errMsg := <-errMsgChan:
				gameUi.ErrorText.SetText(errMsg)
				// ! Ask about Stop Function
				errTimer.Stop()
				errTimer.Reset(time.Second * 3)
			case <-errTimer.C:
				gameUi.ErrorText.SetText("")
			}
		}
	}(mainEnd)

	gameUi.Controller.Start(ctx, nil)
	return err
}
