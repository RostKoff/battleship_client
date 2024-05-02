package main

import (
	"battleship_client/api/client"
	"battleship_client/logic"
)

func main() {
	settings := client.GameSettings{
		AgainstBot: false,
		Nick:       "Rost",
	}

	logic.StartGame(settings)
}
