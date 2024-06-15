package cli

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	wGui "github.com/RostKoff/warships-gui/v2"
)

const (
	fourShip     = "4ship"
	threeShip    = "3ship"
	twoShip      = "2ship"
	oneShip      = "1ship"
	delOpt       = "delete"
	PlacementOpt = "placement"
	GoBack       = "goBack"
)

type shipPlacement struct {
	cluster      [][2]int
	surroundings [][2]int
}

type PlacementUI struct {
	controller   *wGui.GUI
	board        *wGui.Board
	shipsArea    *wGui.HandleArea
	shipsTxt     *wGui.Text
	setShipsBtn  *wGui.Button
	btnsArea     *wGui.HandleArea
	tiles        [10][10]wGui.State
	ships        map[string]Row
	selectedShip string
	shipCoords   []string
}

func InitPlacement(controller *wGui.GUI) *PlacementUI {
	boardCfg := wGui.NewBoardConfig()
	tileCfg := wGui.NewButtonConfig()
	tileCfg.BgColor = boardCfg.ShipColor
	tileCfg.FgColor = boardCfg.TextColor
	tileCfg.Width = 3
	tileCfg.Height = 1
	countCfg := wGui.NewButtonConfig()
	countCfg.BgColor = wGui.Grey
	countCfg.FgColor = wGui.White
	countCfg.Width = 3
	countCfg.Height = 1
	drawables := make([]wGui.Drawable, 0)
	ships := make(map[string]Row, 0)
	handleMap := make(map[string]wGui.Physical, 0)
	x := 50
	y := 4
	shipsTxt := wGui.NewText(x, y-2, "Select the type of ship to place", nil)
	for i := 4; i > 0; i-- {
		tiles := make([]*wGui.Button, i+1)
		countBtn := wGui.NewButton(x, y, fmt.Sprintf("%d", 5-i), countCfg)
		tiles[0] = countBtn
		drawables = append(drawables, countBtn)
		for j := 1; j <= i; j++ {
			button := wGui.NewButton(x+(tileCfg.Width+1)*j, y, string(boardCfg.ShipChar), tileCfg)
			tiles[j] = button
			drawables = append(drawables, button)
		}
		row := NewRow(tiles)
		key := fmt.Sprintf("%dship", i)
		ships[key] = row
		handleMap[key] = row
		y += tileCfg.Height + 1
	}
	{
		tileCfg.Width = 4 + tileCfg.Width*4
		delete := NewRow([]*wGui.Button{wGui.NewButton(x, y, "Delete", tileCfg)})
		key := delOpt
		ships[key] = delete
		handleMap[key] = delete
		drawables = append(drawables, delete.GetButtons()[0])
	}
	states := [10][10]wGui.State{}
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			states[i][j] = wGui.Empty
		}
	}
	setShipsCfg := wGui.NewButtonConfig()
	setShipsCfg.BgColor = wGui.Red
	setShipsBtn := wGui.NewButton(1, 24, "Random configuration", setShipsCfg)
	w, _ := setShipsBtn.Size()
	goBackBtn := wGui.NewButton(2+w, 24, "Go back", setShipsCfg)
	btnsArea := wGui.NewHandleArea(map[string]wGui.Physical{PlacementOpt: setShipsBtn, GoBack: goBackBtn})

	ui := &PlacementUI{
		controller:  controller,
		board:       wGui.NewBoard(2, 2, boardCfg),
		shipsArea:   wGui.NewHandleArea(handleMap),
		shipsTxt:    shipsTxt,
		ships:       ships,
		tiles:       states,
		setShipsBtn: setShipsBtn,
		btnsArea:    btnsArea,
	}

	drawables = append(drawables, ui.board, ui.shipsArea, shipsTxt, ui.btnsArea, ui.setShipsBtn, goBackBtn)
	for _, drawable := range drawables {
		ui.controller.Draw(drawable)
	}
	return ui
}

func (ui *PlacementUI) ShipsSelect(sKey string) (string, error) {
	row, ok := ui.ships[sKey]
	if !ok {
		return "", fmt.Errorf("ship not found")
	}

	btnColor := wGui.Orange
	btnStatus := true
	selectedShip := sKey

	btns := row.GetButtons()
	length := len(btns)
	count := -1
	var err error
	if length > 1 {
		countBtn := btns[0]
		count, err = strconv.Atoi(countBtn.Text())
		if err != nil {
			return "", fmt.Errorf("failed to convert text to integer")
		}
	}

	if sKey == ui.selectedShip || count == 0 {
		btnColor = wGui.Green
		btnStatus = false
		selectedShip = ""
	}

	for i, btn := range btns {
		if i == 0 && len(btns) != 1 {
			continue
		}
		btn.SetBgColor(btnColor)
	}
	for k, clickable := range ui.shipsArea.GetClickables() {
		if k == sKey {
			continue
		}
		clickable.Disabled = btnStatus
	}

	ui.selectedShip = selectedShip
	return ui.selectedShip, nil
}

func (ui *PlacementUI) BoardClick(lCoord, nCoord int, isFirst bool) bool {
	state := ui.tiles[lCoord][nCoord]
	if (state != wGui.Empty && isFirst) || (state != wGui.Emphasis && !isFirst) {
		return false
	}
	ui.tiles[lCoord][nCoord] = wGui.Ship
	ui.changeEmphasisAround(lCoord, nCoord, false)
	ui.board.SetStates(ui.tiles)
	return true
}

func (ui *PlacementUI) Listen(ctx context.Context) error {
	endCtx, end := context.WithCancel(ctx)
	defer end()
	sKey := ui.shipsArea.Listen(ctx)
	sType, err := ui.ShipsSelect(sKey)
	if err != nil {
		return err
	}
	tilesNum := 0
	switch sType {
	case delOpt:
		tilesNum = -1
	case fourShip:
		tilesNum = 4
	case threeShip:
		tilesNum = 3
	case twoShip:
		tilesNum = 2
	case oneShip:
		tilesNum = 1
	}
	innerCtx, canc := context.WithCancel(ctx)
	go func(ctx context.Context) {
		sKey = ui.shipsArea.Listen(ctx)
		select {
		case <-ctx.Done():
			return
		default:
			ui.ShipsSelect(sKey)
			canc()
		}

	}(endCtx)
	if tilesNum == -1 {
		shipNum, err := ui.deleteShip(innerCtx)
		if err != nil {
			return err
		}
		ui.changeShipCounter(fmt.Sprintf("%dship", shipNum), false)
	} else {
		succ := ui.placeShip(tilesNum, innerCtx)
		if succ {
			ui.changeShipCounter(sKey, true)
		}
	}
	if sKey == ui.selectedShip {
		ui.ShipsSelect(sKey)
	}
	if len(ui.shipCoords) == 20 {
		ui.setShipsBtn.SetBgColor(wGui.Green)
		ui.setShipsBtn.SetText("Set configuration")
	} else {
		ui.setShipsBtn.SetBgColor(wGui.Red)
		ui.setShipsBtn.SetText("Random configuration")
	}
	return nil
}

func (ui *PlacementUI) changeShipCounter(sKey string, decrease bool) {
	row, ok := ui.ships[sKey]
	if !ok {
		return
	}
	btns := row.GetButtons()
	if len(btns) == 1 {
		return
	}
	countBtn := btns[0]
	count, err := strconv.Atoi(countBtn.Text())
	if err != nil {
		return
	}

	newCount := 0
	if decrease {
		newCount = count - 1
	} else {
		newCount = count + 1
	}
	countBtn.SetText(fmt.Sprintf("%d", newCount))
}

func (ui *PlacementUI) placeShip(tilesNum int, ctx context.Context) bool {
	if tilesNum == 0 {
		return false
	}
	shipCoords := make([]string, tilesNum)
	var coords [2]int
	var err error
	bCopy := ui.tiles
	for i := 0; i < tilesNum; {
		select {
		case <-ctx.Done():
			ui.board.SetStates(bCopy)
			return false
		default:
			tile := ui.board.Listen(ctx)
			ui.controller.Log(tile)
			shipCoords[i] = tile
			coords, err = ConvertCoords(tile)
			if err != nil {
				continue
			}
			isFirst := false
			if i == 0 {
				isFirst = true
			}

			placed := ui.BoardClick(coords[0], coords[1], isFirst)
			if placed {
				i++
			}
		}
	}
	_, surrodings := getPlacement(coords[0], coords[1], ui.tiles)
	for _, sCoords := range surrodings {
		ui.tiles[sCoords[0]][sCoords[1]] = wGui.Blocked
	}
	ui.board.SetStates(ui.tiles)
	ui.shipCoords = append(ui.shipCoords, shipCoords...)
	return true
}

func (ui *PlacementUI) deleteShip(ctx context.Context) (int, error) {
	for {
		select {
		case <-ctx.Done():
			return 0, nil
		default:
			tile := ui.board.Listen(ctx)
			if tile == "" {
				continue
			}
			coords, err := ConvertCoords(tile)
			if err != nil {
				continue
			}
			state := ui.tiles[coords[0]][coords[1]]
			if state != wGui.Ship {
				continue
			}
			cluster, surroundings := getPlacement(coords[0], coords[1], ui.tiles)
			newShipCoords := make([]string, len(ui.shipCoords))
			bCopy := ui.tiles
			copy(newShipCoords, ui.shipCoords)
			for _, coords := range cluster {
				ship, err := ConvertToString(coords[0], coords[1])
				if err != nil {
					return 0, fmt.Errorf("failed to convert coords: %w", err)
				}
				slices.Sort(newShipCoords)
				pos, found := slices.BinarySearch(newShipCoords, ship)
				if !found {
					continue
				}
				newShipCoords = slices.Delete(newShipCoords, pos, pos+1)
				bCopy[coords[0]][coords[1]] = wGui.Empty
			}
			intersection, err := ui.findIntersection(shipPlacement{cluster: cluster, surroundings: surroundings})
			if err != nil {
				return 0, fmt.Errorf("failed to find intersection: %w", err)
			}
			for _, coords := range surroundings {
				bCopy[coords[0]][coords[1]] = wGui.Empty
			}
			for _, coords := range intersection {
				bCopy[coords[0]][coords[1]] = wGui.Blocked
			}
			ui.shipCoords = newShipCoords
			ui.tiles = bCopy
			ui.board.SetStates(ui.tiles)
			return len(cluster), nil
		}
	}
}

func (ui PlacementUI) findIntersection(ship shipPlacement) ([][2]int, error) {
	ships, err := ui.getShips()
	if err != nil {
		return nil, fmt.Errorf("failed to get ships: %w", err)
	}
	intersection := make([][2]int, 0)
	for _, s := range ships {
		if equalCoordArrays(s.cluster, ship.cluster) {
			continue
		}
		intersection = append(intersection, getIntersection(ship.surroundings, s.surroundings)...)
	}
	return intersection, nil
}

func getIntersection(first, second [][2]int) [][2]int {
	out := make([][2]int, 0)
	bucket := map[[2]int]bool{}

	for _, i := range first {
		for _, j := range second {
			if i == j && !bucket[i] {
				out = append(out, i)
				bucket[i] = true
			}
		}
	}
	return out
}

func (ui PlacementUI) getShips() ([]shipPlacement, error) {
	sCopy := make([]string, len(ui.shipCoords))
	copy(sCopy, ui.shipCoords)
	slices.Sort(sCopy)
	ships := make([]shipPlacement, 0)
	for len(sCopy) != 0 {
		tile := sCopy[0]
		coords, err := ConvertCoords(tile)
		if err != nil {
			return nil, fmt.Errorf("failed to convert coords to num values: %w", err)
		}
		cluster, surroundings := getPlacement(coords[0], coords[1], ui.tiles)
		for _, coords := range cluster {
			tile, err := ConvertToString(coords[0], coords[1])
			if err != nil {
				return nil, fmt.Errorf("failed to convert coords to string: %w", err)
			}
			pos, found := slices.BinarySearch(sCopy, tile)
			if !found {
				continue
			}
			sCopy = slices.Delete(sCopy, pos, pos+1)
		}
		ships = append(ships, shipPlacement{cluster: cluster, surroundings: surroundings})
	}
	return ships, nil
}

func (ui *PlacementUI) Board() []string {
	return ui.shipCoords
}

// func (ui *PlacementUI) deleteShipTile(lCoord, nCoord int) {
// 	cluster, _ := getPlacement(lCoord, nCoord, ui.tiles)
// 	ui.controller.Log("Was here")
// 	ui.tiles[lCoord][nCoord] = wGui.Empty
// 	ui.changeEmphasisAround(lCoord, nCoord, true)

// 	for _, coords := range cluster {
// 		if coords[0] == lCoord && coords[1] == nCoord {
// 			continue
// 		}
// 		ui.changeEmphasisAround(coords[0], coords[1], false)
// 	}
// }

func (ui *PlacementUI) changeEmphasisAround(lCoord, nCoord int, remove bool) {
	var oldState, newState wGui.State
	if remove {
		oldState = wGui.Emphasis
		newState = wGui.Empty
	} else {
		oldState = wGui.Empty
		newState = wGui.Emphasis
	}

	for i := -1; i < 2; i += 2 {
		newLCoord := lCoord + i
		newNCoord := nCoord + i
		if newLCoord >= 0 && newLCoord <= 9 {
			ui.replaceTile(newLCoord, nCoord, oldState, newState)
		}
		if newNCoord >= 0 && newNCoord <= 9 {
			ui.replaceTile(lCoord, newNCoord, oldState, newState)
		}
	}
}

func (ui *PlacementUI) replaceTile(lCoord, nCoord int, old, new wGui.State) {
	state := ui.tiles[lCoord][nCoord]
	if state != old {
		return
	}
	ui.tiles[lCoord][nCoord] = new
}

func getPlacement(lCoord, nCoord int, states [10][10]wGui.State) (statusCluster, surroundings [][2]int) {
	toVisit := [][2]int{{lCoord, nCoord}}
	statusCluster = make([][2]int, 0)
	surroundings = make([][2]int, 0)
	for length := len(toVisit); len(toVisit) > 0; length = len(toVisit) {
		n := length - 1
		coords := toVisit[n]
		x := coords[0]
		y := coords[1]
		toVisit = toVisit[:n]
		state := states[x][y]

		included := false
		for _, c := range statusCluster {
			if c[0] == coords[0] && c[1] == coords[1] {
				included = true
				break
			}
		}
		if !included {
			statusCluster = append(statusCluster, coords)
		}
		for i := -1; i < 2; i++ {
			for j := -1; j < 2; j++ {
				if (j == 0 && i == 0) || x+i < 0 || x+i > 9 || y+j < 0 || y+j > 9 {
					continue
				}
				nearState := states[x+i][y+j]
				if nearState == state {
					c := [2]int{x + i, y + j}
					if !slices.Contains(statusCluster, c) {
						toVisit = append(toVisit, c)
					}
					continue
				}
				surroundings = append(surroundings, [2]int{x + i, y + j})
			}
		}
	}
	return
}

func equalCoordArrays(first, second [][2]int) bool {
	for _, elem := range first {
		elemPresent := false
		for _, elem2 := range second {
			if elem[0] == elem2[0] && elem[1] == elem2[1] {
				elemPresent = true
			}
		}
		if !elemPresent {
			return false
		}
	}
	return true
}

func (ui *PlacementUI) SetBtnListen(ctx context.Context) string {
	return ui.btnsArea.Listen(ctx)
}

func (ui *PlacementUI) ShipCoords() []string {
	return ui.shipCoords
}
