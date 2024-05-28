package logic

import (
	"battleship_client/api/client"
	"battleship_client/gui/cli"
	"context"

	wGui "github.com/RostKoff/warships-gui/v2"
)

func DisplayGameSettings(gui *wGui.GUI, ch chan<- client.GameSettings) {
	refresh := make(chan rune)

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
	x, _ = botBtn.Position()
	w, _ = botBtn.Size()
	btnCfg.BgColor = wGui.Grey
	refreshBtn := wGui.NewButton(x+w+2, 11, "Refresh", btnCfg)

	// Handle Area for buttons
	btnMapping := map[string]wGui.Physical{
		"botBtn":     botBtn,
		"startBtn":   startBtn,
		"refreshBtn": refreshBtn,
	}
	btnArea := wGui.NewHandleArea(btnMapping)

	// Lobby
	lCfg := wGui.NewButtonConfig()
	lCfg.Width = 42

	lobbyTxt := wGui.NewButton(2, 15, "Lobby", lCfg)
	lCfg.Width = 21
	statusHead := wGui.NewButton(2, 18, "Status", lCfg)
	nickHead := wGui.NewButton(lCfg.Width+2, 18, "Nick", lCfg)

	drawables := []wGui.Drawable{
		nameTxt, nameIn, descTxt, descIn, btnArea, botBtn, startBtn, lobbyTxt, statusHead, nickHead, refreshBtn,
	}

	areaMap := make(map[string]wGui.Physical)
	lobbyArea := wGui.NewHandleArea(areaMap)
	lobbyMap := make(map[string]cli.Row)
	go func() {
		for {
			lobbyMap = getLobbyGames()
			for key, value := range lobbyMap {
				areaMap[key] = value
				btns := value.GetButtons()
				gui.Draw(btns[0])
				gui.Draw(btns[1])
			}
			lobbyArea.SetClickablesOn(areaMap)
			gui.Draw(lobbyArea)
			<-refresh
			gui.Remove(lobbyArea)
			for key, value := range lobbyMap {
				btns := value.GetButtons()
				gui.Remove(btns[0])
				gui.Remove(btns[1])
				delete(areaMap, key)
			}
		}
	}()

	// Draw objects
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
		case "refreshBtn":
			targetNick = ""
			startBtn.SetBgColor(wGui.Green)
			startBtn.SetText("Host Game")
			targetRow = nil
			refresh <- 'r'
		}
	}
}

func getLobbyGames() map[string]cli.Row {
	lobbyGames, err := client.Lobby()
	lCfg := wGui.NewButtonConfig()
	lCfg.Width = 21
	if err != nil {
		return nil
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
	}
	return lobbyMap
}
