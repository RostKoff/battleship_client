package main

import (
	"fmt"
	"warships_client/client"
)

func main() {
	settings := client.GameSettings{
		Nick:        "custom_nick",
		Description: "Custom desc",
		AgainstBot:  true,
	}

	c, err := client.InitGame(settings)
	if err != nil {
		panic(err)
	}
	board, _ := client.Board(c)
	fmt.Println(board)
}
