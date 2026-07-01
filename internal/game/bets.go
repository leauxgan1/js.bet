package game

var Bets map[string]int = make(map[string]int)

func SetBet(name string, amount int) {
	Bets[name] = amount
}

func ClearBets() {
	for k := range Bets {
		delete(Bets, k)
	}
}
