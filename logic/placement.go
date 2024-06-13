package logic

import (
	"battleship_client/gui/cli"
	"context"

	wGui "github.com/RostKoff/warships-gui/v2"
)

func DisplayPlacement(controller *wGui.GUI, placement chan<- []string, abort chan<- rune) {
	controller.NewScreen("placement")
	controller.SetScreen("placement")
	defer controller.RemoveScreen("placement")
	ui := cli.InitPlacement(controller)
	ctx, mainEnd := context.WithCancel(context.Background())
	defer mainEnd()
	go handlePlacementClick(ui, ctx)
	opt := ui.SetBtnListen(ctx)
	switch opt {
	case cli.PlacementOpt:
		coords := ui.ShipCoords()
		if len(coords) == 20 {
			placement <- ui.ShipCoords()
		} else {
			placement <- []string{}
		}
	case cli.GoBack:
		abort <- ' '
	}
}

func handlePlacementClick(ui *cli.PlacementUI, ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		for {
			ui.Listen(ctx)
		}
	}
}
