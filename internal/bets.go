package internal

import (
	"js-bet/internal/game"
	"log"
)

type BetDetails struct {
	BetAmount int
	BetSide   bool // (true, false) => (Left, Right)
}

var Bets map[string]BetDetails = make(map[string]BetDetails, 10)

func SetBet(name string, amount int, side bool) {
	Bets[name] = BetDetails{amount, side}
	var sideStr string
	if side {
		sideStr = "Left"
	} else {
		sideStr = "Right"
	}
	log.Printf("%s Bet on %s with an amount of %d", name, sideStr, amount)
}

func ClearBets() {
	for k := range Bets {
		delete(Bets, k)
	}
}

func AwardBet(name string, winnerResult game.WinnerEnum) {
	details := Bets[name]
	if details.BetSide == true && winnerResult == game.LEFT {
		// Award money to this player
	} else if details.BetSide == false && winnerResult == game.RIGHT {
		// Award money
	}
	// Give nothing
}

func AwardBets(winner game.WinnerEnum) {
	log.Print(sseHub.clients)
	// For each name in our map of Bets, award that user with double the amount they put in if they succeeded.
	// Otherwise, reduce their gold by the amount they bet
	// for key, val := range Bets {

	// }
	// Then, for each client currently connected, send them an html update of their client state
}
