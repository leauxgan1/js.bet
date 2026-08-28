package internal

import (
	"js-bet/internal/game"
	// "log"
)

type BetDetails struct {
	BetAmount int
	BetSide   bool // (true, false) => (Left, Right)
}

var Bets map[int64]BetDetails = make(map[int64]BetDetails, 10)

func SetBet(userId int64, amount int, side bool) {
	Bets[userId] = BetDetails{amount, side}
	// var sideStr string
	// if side {
	// 	sideStr = "Left"
	// } else {
	// 	sideStr = "Right"
	// }
	// log.Printf("User with ID: %d Bet on %s with an amount of %d", userId, sideStr, amount)
}

func ClearBets() {
	for k := range Bets {
		delete(Bets, k)
	}
}

func awardBet(client Client, userId int64, details BetDetails, winnerResult game.WinnerEnum) {
	if details.BetSide == true && winnerResult == game.LEFT || details.BetSide == false && winnerResult == game.RIGHT {
		// Award money to this player
		db.ChangeUserGold(userId, 2*details.BetAmount)
	} else {
		db.ChangeUserGold(userId, -details.BetAmount)
	}
	updateGoldGUI(client) // Send an update to that client since their gold amount definitely changed
}

func updateGoldGUI(client Client) {

}

func AwardBets(winner game.WinnerEnum) {
	// log.Print(sseHub.clients)

	// For each name in our map of Bets, award that user with double the amount they put in if they succeeded.
	// Otherwise, reduce their gold by the amount they bet
	for userId, betDetails := range Bets {
		client := userClientMap[userId]
		awardBet(client, userId, betDetails, winner)
	}
	// Send out new HTML fragment for each user connected with a matching ID
}
