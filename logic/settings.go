package logic

import (
	"battleship_client/api/client"
	"battleship_client/gui/cli"
	"context"
	"fmt"

	wGui "github.com/RostKoff/warships-gui/v2"
)

// Displays game settings and listens for button clicks.
// When the start button or bot button is clicked it sends game settings to the channel given as the argument.
func DisplayGameSettings(controller *wGui.GUI, ch chan<- client.GameSettings) {
	refresh := make(chan rune)

	controller.NewScreen("settings")
	controller.SetScreen("settings")

	settingsUi := cli.InitSettings(controller)

	go displayLobby(settingsUi, refresh)

	ctx := context.Background()
	go handleLobby(settingsUi, ctx)

	// Handle button clicks.
	for {
		clicked := settingsUi.BtnArea.Listen(ctx)
		switch clicked {
		case "botBtn":
			ch <- client.GameSettings{
				AgainstBot:  true,
				Nick:        settingsUi.Nick(),
				Description: settingsUi.Desc(),
			}
			return
		case "startBtn":
			ch <- client.GameSettings{
				AgainstBot:  false,
				Nick:        settingsUi.Nick(),
				Description: settingsUi.Desc(),
				TargetNick:  settingsUi.TargetNick(),
			}
			return
		case "refreshBtn":
			settingsUi.ToggleOpponent(settingsUi.TargetNick())
			refresh <- 'r'
		}
	}
}

// Fetches game lobbies from the API, displays them on the screen and waits until any rune is sent to the channel given as the argument.
func displayLobby(ui *cli.SettingsUI, refresh <-chan rune) {
	for {
		lobbyGames, err := client.Lobby()
		if err != nil {
			ui.Controller.Log(fmt.Sprintf("failed to get game lobby: %s", err.Error()))
			lobbyGames = nil
		}
		ui.DrawLobbyGames(lobbyGames)
		<-refresh
	}
}

func handleLobby(ui *cli.SettingsUI, ctx context.Context) {
	for {
		ui.ListenLobby(ctx)
	}
}
