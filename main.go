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
		fmt.Printf("%s\n%v\n", board, status)
		fmt.Scanln(&line)
		res, err := g.Fire(line)
		if err != nil {
			panic(err)
		}
		fmt.Println(res)
	}
}
