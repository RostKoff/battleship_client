package logic

import (
	"battleship_client/api/client"
	"battleship_client/gui/cli"
	"context"
	"fmt"
	"time"

	wGui "github.com/RostKoff/warships-gui/v2"
)

func DisplayGameSettings(gui *wGui.GUI, ch chan<- client.GameSettings) {
	gui.NewScreen("settings")
	gui.SetScreen("settings")

	// Name Input
	nameTxt := wGui.NewText(2, 1, "Enter name", nil)
	nameInCfg := wGui.NewTextFieldConfig()
	nameInCfg.UnfilledChar = '_'
	nameInCfg.InputOn = true
	nameIn := wGui.NewTextField(2, 2, 30, 1, nameInCfg)

	// Description Input
	descTxt := wGui.NewText(2, 4, "Enter Description", nil)
	descInCfg := wGui.NewTextFieldConfig()
	descInCfg.UnfilledChar = '.'
	descInCfg.InputOn = true
	descIn := wGui.NewTextField(2, 5, 30, 5, descInCfg)

	// Action Buttons
	btnCfg := wGui.NewButtonConfig()
	btnCfg.Width = 0
	btnCfg.Height = 0
	btnCfg.Width = 20
	btnCfg.BgColor = wGui.Green
	startBtn := wGui.NewButton(2, 11, "Host Game", btnCfg)
	x, _ := startBtn.Position()
	w, _ := startBtn.Size()
	btnCfg.BgColor = wGui.Blue
	botBtn := wGui.NewButton(x+w+2, 11, "Against Bot", btnCfg)

	// Handle Area for buttons
	btnMapping := map[string]wGui.Physical{
		"botBtn":   botBtn,
		"startBtn": startBtn,
	}
	btnArea := wGui.NewHandleArea(btnMapping)

	// Lobby
	lCfg := wGui.NewButtonConfig()
	lCfg.Width = 42

	lobbyTxt := wGui.NewButton(2, 15, "Lobby", lCfg)
	lobbyGames, err := client.Lobby()
	if err != nil {
		return
	}
	lCfg.Width = 21
	statusHead := wGui.NewButton(2, 18, "Status", lCfg)
	nickHead := wGui.NewButton(lCfg.Width+2, 18, "Nick", lCfg)

	drawables := []wGui.Drawable{
		nameTxt, nameIn, descTxt, descIn, btnArea, botBtn, startBtn, lobbyTxt, statusHead, nickHead,
	}

	lobbyMap := make(map[string]cli.Row, 0)
	for i, game := range lobbyGames {
		y := 18 + (i+1)*3
		btns := []*wGui.Button{
			wGui.NewButton(2, y, game.Status, lCfg),
			wGui.NewButton(lCfg.Width+2, y, game.Nick, lCfg),
		}

		row := cli.NewRow(btns)
		lobbyMap[game.Nick] = row
		drawables = append(drawables, btns[0], btns[1])
	}

	areaMap := make(map[string]wGui.Physical)
	for key, value := range lobbyMap {
		areaMap[key] = value
	}
	lobbyArea := wGui.NewHandleArea(areaMap)

	// Draw objects
	drawables = append(drawables, lobbyArea)
	for _, drawable := range drawables {
		gui.Draw(drawable)
	}

	ctx := context.Background()
	targetNick := ""
	targetRow := new(cli.Row)
	go func() {
		for {
			clickedNick := lobbyArea.Listen(ctx)
			row, ok := lobbyMap[clickedNick]
			if !ok {
				continue
			}
			if targetNick == clickedNick {
				row.SetBgColor(wGui.Black)
				row.SetFgColor(wGui.White)
				targetNick = ""
				startBtn.SetBgColor(wGui.Green)
				startBtn.SetText("Host Game")
				targetRow = nil
				continue
			}
			row.SetBgColor(wGui.White)
			row.SetFgColor(wGui.Black)
			targetNick = clickedNick
			startBtn.SetBgColor(wGui.Red)
			startBtn.SetText("Fight Opponent")
			if targetRow != nil {
				targetRow.SetBgColor(wGui.Black)
				targetRow.SetFgColor(wGui.White)
			}
			targetRow = &row
		}
	}()

	for {
		clicked := btnArea.Listen(ctx)
		switch clicked {
		case "botBtn":
			ch <- client.GameSettings{
				AgainstBot:  true,
				Nick:        nameIn.GetText(),
				Description: descIn.GetText(),
			}
		case "startBtn":
			ch <- client.GameSettings{
				AgainstBot:  false,
				Nick:        nameIn.GetText(),
				Description: descIn.GetText(),
				TargetNick:  targetNick,
			}
		}
	}
}

func StartGame(controller *wGui.GUI, gs client.GameSettings) error {
	controller.NewScreen("game")
	controller.SetScreen("game")
	waitTxt := wGui.NewText(1, 1, "Waiting for game to start...", nil)
	controller.Draw(waitTxt)

	statusRes := client.StatusResponse{}
	apiClient, err := client.InitGame(gs)
	if err != nil {
		return fmt.Errorf("failed to initialise the game, %w", err)
	}

	// The channel will send messages to the goroutine responsible for displaying errors.
	errMsgChan := make(chan string)

	// Requesting the API for the status of the game until it is started.
	for refreshCount := 0; ; refreshCount++ {
		if refreshCount == 10 {
			err := apiClient.Refresh()
			if err != nil {
				return fmt.Errorf("failed to refresh game session: %w", err)
			}
			refreshCount = 0
		}
		statusRes, err = apiClient.Status()
		if err != nil {
			return fmt.Errorf("failed to get game status: %w", err)
		}
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
	controller.Remove(waitTxt)
	gameUi := cli.InitGameUI(controller)
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
	gameUi.PlaceAndDrawDescriptions(pDesc, oppDesc)

	// Context to cancel additional goroutines after game is finished.
	mainEnd, cancel := context.WithCancel(context.Background())
	// Goroutine that is responsible for updating GUI according to the game status got from API.
	defer cancel()

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
					errMsgChan <- "Failed to fire!"
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
				errTimer.Stop()
				errTimer.Reset(time.Second * 3)
			case <-errTimer.C:
				gameUi.ErrorText.SetText("")
			}
		}
	}(mainEnd)

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
