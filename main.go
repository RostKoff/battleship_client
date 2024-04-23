package main

import (
	"battleship_client/client"
	"fmt"
)

func main() {
	settings := client.GameSettings{
		Nick:        "custom_nick",
		Description: "Custom desc",
		AgainstBot:  true,
	}

	g, err := client.InitGame(settings)
	if err != nil {
		panic(err)
	}
	line := ""
	for {
		board, _ := g.Board()
		status, _ := g.Status()
		descs, err := g.PlayerDescriptions()
		if err != nil {
			fmt.Printf("Desc error: %s", err)
		}
		fmt.Printf("board: %s\nstatus: %v\nplayer desc: %s\topp desc: %s\n", board, status, descs.PlayerDescription, descs.OpponentDescription)
		fmt.Scanln(&line)
		res, err := g.Fire(line)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}
}
