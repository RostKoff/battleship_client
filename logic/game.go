package logic

import (
	"battleship_client/api/client"
	"battleship_client/gui/cli"
	"context"
	"fmt"
	"time"

	wGui "github.com/RostKoff/warships-gui/v2"
)

func StartGame(controller *wGui.GUI, gs client.GameSettings, abandon chan<- rune) error {
	controller.NewScreen("game")
	controller.SetScreen("game")

	apiClient, err := client.InitGame(gs)
	if err != nil {
		return fmt.Errorf("failed to initialise the game, %w", err)
	}

	statusRes, err := waitUntilStart(apiClient, controller)
	if err != nil {
		return fmt.Errorf("fail occured while waiting for start: %w", err)
	}
	gameUi, err := displayGame(apiClient, controller, statusRes)
	if err != nil {
		return fmt.Errorf("failed to display the game: %w", err)
	}

	// Context to cancel additional goroutines after game is finished.
	mainEnd, cancel := context.WithCancel(context.Background())
	// Goroutine that is responsible for updating GUI according to the game status got from API.
	defer cancel()

	// The channel will send messages to the goroutine responsible for displaying errors.
	errMsgChan := make(chan string)

	go btnListen(mainEnd, gameUi, apiClient, abandon)

	go errorDisplayer(mainEnd, gameUi, errMsgChan)

	go handleShot(mainEnd, gameUi, apiClient, errMsgChan)

	board, err := apiClient.Board()
	if err != nil {
		return fmt.Errorf("failed to get player's ship location: %w", err)
	}
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
				return err
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
		return err
	}
	gameUi.EndText.SetText("You won!\n")
	// Goroutine responsible for getting player input and handling fire mechanic.
	return err
}

// Fetches the status from the API every second until the game starts, and refreshes the game session every 10 seconds.
// Returns the status of the started game.
func waitUntilStart(apiClient client.GameClient, controller *wGui.GUI) (statusRes client.StatusResponse, err error) {
	waitTxt := wGui.NewText(1, 1, "Waiting for game to start...", nil)
	controller.Draw(waitTxt)

	// Requesting the API for the status of the game until it is started.
	// Refresh count is used to refresh game session every 10 seconds.
	for refreshCount := 0; ; refreshCount++ {
		if refreshCount == 10 {
			err = apiClient.Refresh()
			if err != nil {
				return statusRes, fmt.Errorf("failed to refresh game session: %w", err)
			}
			refreshCount = 0
		}
		statusRes, err = apiClient.Status()
		if err != nil {
			return statusRes, fmt.Errorf("failed to get game status: %w", err)
		}
		if statusRes.Status != "game_in_progress" {
			time.Sleep(time.Second)
			continue
		}
		break
	}
	controller.Remove(waitTxt)
	return statusRes, nil
}

func displayGame(apiClient client.GameClient, controller *wGui.GUI, statusRes client.StatusResponse) (gameUi *cli.GameUI, err error) {
	board, err := apiClient.Board()
	if err != nil {
		return nil, fmt.Errorf("failed to get player's ship location: %w", err)
	}
	gameUi = cli.InitGameUI(controller)
	// Fill the board with ships
	for _, coord := range board {
		gameUi.PBoard.UpdateState(coord, wGui.Ship)
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
	gameUi.DrawDescriptions(pDesc, oppDesc)
	return
}

// Responsible for logic related to the shot. The context is used to end the function when the game is over.
func handleShot(ctx context.Context, gameUi *cli.GameUI, client client.GameClient, errChan chan<- string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			coord, err := gameUi.ListenForShot(ctx)
			if err != nil {
				errChan <- "Failed to handle click!"
				gameUi.Controller.Log(fmt.Sprintf("Listen Error: %s", err.Error()))
			}
			fireRes, err := client.Fire(coord)
			if err != nil {
				errChan <- "Failed to fire!"
				gameUi.Controller.Log(fmt.Sprintf("Fire error: %s", err.Error()))
				continue
			}
			err = gameUi.HandlePShot(fireRes, coord)
			if err != nil {
				errChan <- "Failed to handle player shot"
				gameUi.Controller.Log(fmt.Sprintf("Player shot error: %s", err.Error()))
				continue
			}
			gameUi.CalculateAccuracy()
		}
	}
}

// Displays an error message received from the `errChan` for 3 seconds and then hides it.
func errorDisplayer(ctx context.Context, gameUi *cli.GameUI, errChan <-chan string) {
	// Initilise the timer.
	errTimer := time.NewTimer(time.Second * 10)
	errTimer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case errMsg := <-errChan:
			// Displays new message and resets the timer.
			gameUi.ErrorText.SetText(errMsg)
			errTimer.Stop()
			errTimer.Reset(time.Second * 3)
		case <-errTimer.C:
			// Hides the message when the timer expires.
			gameUi.ErrorText.SetText("")
		}
	}
}

func btnListen(ctx context.Context, gameUi *cli.GameUI, client client.GameClient, abandon chan<- rune) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			opt := gameUi.BtnListen(ctx)
			switch opt {
			case cli.AbandonOpt:
				client.Abandon()
				abandon <- ' '
				return
			}
		}
	}
}
