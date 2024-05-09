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
	if err != nil {
		return fmt.Errorf("failed to initialise the game, %w", err)
	}

	// The channel will send messages to the goroutine responsible for displaying errors.
	errMsgChan := make(chan string)

	// Requesting the API for the status of the game until it is started.
	// Returns an error if the status request failed more times than specified in `maxNumErr`.
	maxErrNum := 5
	for errCount := 0; ; {
		statusRes, err = apiClient.Status()
		if err != nil {
			if errCount > maxErrNum {
				return fmt.Errorf("failed to get game status: %w", err)
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
		return fmt.Errorf("failed to get player board: %w", err)
	}
	// Fill the board with ships
	for _, coord := range board {
		gameUi.PBoard.UpdateState(coord, warships_gui.Ship)
	}
	gameUi.DrawNicks(statusRes.Nick, statusRes.Opponent)

	var pDesc, oppDesc string
	descs, err := apiClient.PlayerDescriptions()
	if err != nil {
		gameUi.Controller.Log("Player Descriptions Error: %s", err)
		pDesc = "n/a"
		oppDesc = "n/a"
	} else {
		pDesc = descs.PlayerDescription
		oppDesc = descs.OpponentDescription
	}
	gameUi.PlaceAndDrawDescriptions(pDesc, oppDesc)

	// Context to cancel additional goroutines after game is finished.
	mainEnd, cancel := context.WithCancel(context.Background())
	// Goroutine that is responsible for updating GUI according to the game status got from API.
	go func() {
		defer cancel()
		oppShotCount := 0
		// Main game loop. Gets the game status every second and updates the GUI accordingly
		for {
			statusRes, err = apiClient.Status()
			if err != nil {
				gameUi.Controller.Log(fmt.Sprintf("Status error: %s", err.Error()))
				errMsgChan <- "Failed to get game status"
				time.Sleep(time.Second)
				continue
			}
			// Updates Player's board to display the opponent's shots, but only if there are new ones.
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
					errMsgChan <- err.Error()
					gameUi.Controller.Log(fmt.Sprintf("Fire error: %s", err.Error()))
					continue
				}
				err = gameUi.HandlePShot(fireRes, coord)
				if err != nil {
					errMsgChan <- "Failed to handle player shot"
					gameUi.Controller.Log(fmt.Sprintf("Player shot error: %s", err.Error()))
					continue
				}
			}
		}
	}(mainEnd)

	// Displays and error message for 3 seconds and then hides it.
	go func(ctx context.Context) {
		// Initilise the timer.
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

	ctx := context.Background()
	gameUi.Controller.Start(ctx, nil)
	return err
}
