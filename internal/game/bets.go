package game

import "log"

var Bets map[string]int = make(map[string]int)

func SetBet(name string, amount int) {
	Bets[name] = amount
	log.Printf("Bets updated to: %v", Bets)
}

func ClearBets() {
	for k := range Bets {
		delete(Bets, k)
	}
}
