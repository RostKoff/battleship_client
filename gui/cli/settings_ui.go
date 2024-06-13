package cli

import (
	"battleship_client/api/client"
	"context"

	wGui "github.com/RostKoff/warships-gui/v2"
)

type SettingsUI struct {
	Controller *wGui.GUI
	nameInput  *wGui.TextField
	descInput  *wGui.TextField
	startBtn   *wGui.Button
	botBtn     *wGui.Button
	refreshBtn *wGui.Button
	BtnArea    *wGui.HandleArea
	lobbyArea  *wGui.HandleArea
	lobbyMap   map[string]Row
	targetNick string
	targetRow  *Row
}

// Creates and draws all the elements of the lobby.
func InitSettings(controller *wGui.GUI, settings *client.GameSettings) *SettingsUI {
	// Name Input
	nameTxt := wGui.NewText(2, 1, "Enter name", nil)
	nameInCfg := wGui.NewTextFieldConfig()
	nameInCfg.UnfilledChar = '_'
	nameInCfg.InputOn = true
	nameIn := wGui.NewTextField(2, 2, 30, 1, nameInCfg)
	nameIn.SetText(settings.Nick)

	// Description Input
	descTxt := wGui.NewText(2, 4, "Enter Description", nil)
	descInCfg := wGui.NewTextFieldConfig()
	descInCfg.UnfilledChar = '_'
	descInCfg.InputOn = true
	descIn := wGui.NewTextField(2, 5, 30, 5, descInCfg)
	descIn.SetText(settings.Description)

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
	nickHead := wGui.NewButton(2, 18, "Nick", lCfg)
	statusHead := wGui.NewButton(lCfg.Width+2, 18, "Status", lCfg)
	lobbyArea := wGui.NewHandleArea(nil)

	// Draw all objects
	drawables := []wGui.Drawable{
		nameTxt,
		nameIn,
		descTxt,
		descIn,
		startBtn,
		botBtn,
		refreshBtn,
		btnArea,
		lobbyTxt,
		lobbyArea,
		nickHead,
		statusHead,
	}
	for _, drawable := range drawables {
		controller.Draw(drawable)
	}

	return &SettingsUI{
		Controller: controller,
		nameInput:  nameIn,
		descInput:  descIn,
		startBtn:   startBtn,
		botBtn:     botBtn,
		refreshBtn: refreshBtn,
		BtnArea:    btnArea,
		lobbyArea:  lobbyArea,
	}
}

// Creates a row for each given lobby game and displays them.
// Also adds them to the `lobbyArea` handle area, making it possible to listen for clicks.
func (ui *SettingsUI) DrawLobbyGames(lobbyGames []client.LobbyGame) {
	ui.clearLobby()

	lCfg := wGui.NewButtonConfig()
	lCfg.Width = 21
	ui.lobbyMap = make(map[string]Row, 0)
	areaMap := make(map[string]wGui.Physical)
	for i, game := range lobbyGames {
		y := 18 + (i+1)*3
		btns := []*wGui.Button{
			wGui.NewButton(2, y, game.Nick, lCfg),
			wGui.NewButton(lCfg.Width+2, y, game.Status, lCfg),
		}
		ui.Controller.Draw(btns[0])
		ui.Controller.Draw(btns[1])
		row := NewRow(btns)
		areaMap[game.Nick] = row
		ui.lobbyMap[game.Nick] = row
	}
	ui.lobbyArea.SetClickablesOn(areaMap)
	ui.Controller.Draw(ui.lobbyArea)
}

// Removes all the lobby rows and their handle area from the screen.
func (ui *SettingsUI) clearLobby() {
	ui.Controller.Remove(ui.lobbyArea)
	for _, row := range ui.lobbyMap {
		btns := row.GetButtons()
		ui.Controller.Remove(btns[0])
		ui.Controller.Remove(btns[1])
	}
}

// Listens for a click in the lobby, and then toggles the row that was clicked.
func (ui *SettingsUI) ListenLobby(ctx context.Context) {
	clickedNick := ui.lobbyArea.Listen(ctx)
	ui.ToggleOpponent(clickedNick)
}

// It is responsible for the possibility to select an opponent and changes the graphical interface depending on the given nickname of the opponent.
func (ui *SettingsUI) ToggleOpponent(nick string) {
	// If nick is not present in the lobby - end.
	row, ok := ui.lobbyMap[nick]
	if !ok {
		return
	}
	// If the nickname is the same as the already selected opponent, deselect it.
	if ui.targetNick == nick {
		row.SetBgColor(wGui.Black)
		row.SetFgColor(wGui.White)
		ui.targetNick = ""
		ui.startBtn.SetBgColor(wGui.Green)
		ui.startBtn.SetText("Host Game")
		ui.targetRow = nil
		return
	}
	// else - select given opponent.
	row.SetBgColor(wGui.White)
	row.SetFgColor(wGui.Black)
	ui.targetNick = nick
	ui.startBtn.SetBgColor(wGui.Red)
	ui.startBtn.SetText("Fight Opponent")
	// If another opponent was selected before, change the colour of their row to normal.
	if ui.targetRow != nil {
		ui.targetRow.SetBgColor(wGui.Black)
		ui.targetRow.SetFgColor(wGui.White)
	}
	ui.targetRow = &row
}

func (ui *SettingsUI) TargetNick() string {
	return ui.targetNick
}

func (ui *SettingsUI) Nick() string {
	return ui.nameInput.GetText()
}

func (ui *SettingsUI) Desc() string {
	return ui.descInput.GetText()
}
