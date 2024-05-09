package main

import (
	"battleship_client/api/client"
	"battleship_client/logic"
	"fmt"
)

func main() {
	settings := client.GameSettings{
		AgainstBot:  true,
		Nick:        "Zamog",
		Description: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	}
	err := logic.StartGame(settings)
	if err != nil {
		fmt.Println(err)
	}

	// splitted := cli.SplitWord(3, "AAAAAAA")
	// fmt.Println(splitted)
}
