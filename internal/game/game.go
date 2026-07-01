package game

import (
	"fmt"
	"js-bet/internal/eventlog"
	"log"
	"math/rand/v2"
)

type Action struct {
	Direction InitiativeEnum
	WasHit    bool
	WasCrit   bool
}

type GameState struct {
	LeftFighter  Fighter
	RightFighter Fighter
	Winner       WinnerEnum
	FrameCount   int
	LastAction   Action
	AudioPlayers AudioPlayer
}

func New() GameState {
	leftFighter := chooseRandomFighter()
	rightFighter, err := chooseRandomFighterExclusive(leftFighter.Name)
	if err != nil {
		log.Panic(err)
	}
	return GameState{
		LeftFighter:  leftFighter,
		RightFighter: rightFighter,
		Winner:       NEITHER,
	}
}

func (g *GameState) ResetKeepWinner() {
	switch g.Winner {
	case LEFT:
		g.LeftFighter.Reset()
		newRight, err := chooseRandomFighterExclusive(g.LeftFighter.Name)
		if err != nil {
			return
		}
		g.RightFighter = newRight
	case RIGHT:
		g.RightFighter.Reset()
		newLeft, err := chooseRandomFighterExclusive(g.RightFighter.Name)
		if err != nil {
			return
		}
		g.LeftFighter = newLeft
	default:
		*g = New()
	}
}

type WinnerEnum uint

const (
	_ = iota
	LEFT
	RIGHT
	NEITHER
)

type InitiativeEnum uint

const (
	LEFT_TO_RIGHT = 0
	RIGHT_TO_LEFT = 1
)

func (g *GameState) Act(initiative InitiativeEnum) {
	// Determine acting direction based on initiative enum
	var fighter *Fighter
	var oppFighter *Fighter
	if initiative == LEFT_TO_RIGHT {
		fighter = &g.LeftFighter
		oppFighter = &g.RightFighter
	} else { // RIGHT_TO_LEFT
		fighter = &g.RightFighter
		oppFighter = &g.LeftFighter
	}

	// Action logic

	fighter.Timer.Value = fighter.Timer.MaxValue // Reset timer

	hit := fighter.CheckHit()
	damage := fighter.Damage.Value
	fighter.State = ATTACKING
	if !hit {
		g.AudioPlayers.DodgePlaying = true
		eventlog.EventLog.Write(fmt.Sprintf("%s just missed...", fighter.Name))
		return
	} else {
		g.AudioPlayers.AttackPlaying = true
	}
	oppFighter.State = DEFENDING
	g.AudioPlayers.BlockPlaying = true
	crit := fighter.CheckCrit()
	if crit {
		g.AudioPlayers.AttackPlaying = false
		g.AudioPlayers.CritPlaying = true
		damage *= 2.0
		fighter.State = CRITTING
		eventlog.EventLog.Write(fmt.Sprintf("%s just critically hit %s for %d", fighter.Name, oppFighter.Name, damage))
	} else {
		eventlog.EventLog.Write(fmt.Sprintf("%s just hit %s for %d", fighter.Name, oppFighter.Name, damage))
	}
	oppFighter.Health.Value -= damage
}

func (f Fighter) CheckHit() bool {
	if f.Accuracy.Value > 0.0 && rand.Float32() < f.Accuracy.Value {
		return true
	}
	return false
}

func (f Fighter) CheckCrit() bool {
	if f.CritRate.Value > 0.0 && rand.Float32() < f.CritRate.Value {
		return true
	}
	return false
}

// TODO add sound for winner being determined
func determineWinner(left Fighter, right Fighter) WinnerEnum {
	if left.Health.Value <= 0 && right.Health.Value <= 0 {
		if left.Health.Value < right.Health.Value {
			return RIGHT
		} else {
			return LEFT
		}
	} else if left.Health.Value <= 0 {
		return RIGHT
	} else if right.Health.Value <= 0 {
		return LEFT
	}
	return NEITHER
}

func (g *GameState) StepGame() {
	g.FrameCount += 1
	g.RightFighter.State = READY
	g.LeftFighter.State = READY
	g.AudioPlayers.Stop()

	// Check for a non-positive health, choose a winner and keep them in the game for the next round
	var winner = determineWinner(g.LeftFighter, g.RightFighter)
	if winner == LEFT || winner == RIGHT {
		g.Winner = winner
		g.ResetKeepWinner()
		return
	}

	lReady := g.LeftFighter.Timer.Value <= 0
	rReady := g.RightFighter.Timer.Value <= 0
	if lReady && rReady {
		if g.LeftFighter.Timer.Value == g.RightFighter.Timer.Value { // Choose lesser Timer when both ready, higher speed on ties
			if g.LeftFighter.Speed.Value == g.RightFighter.Speed.Value {
				rand := rand.Float32() // Choose randomly on second tie
				if rand < 0.5 {        // Left fighter acts
					g.Act(LEFT_TO_RIGHT)
				} else { // Right fighter acts
					g.Act(RIGHT_TO_LEFT)
				}
			} else if g.LeftFighter.Speed.Value > g.RightFighter.Speed.Value {
				g.Act(LEFT_TO_RIGHT)
			} else {
				g.Act(RIGHT_TO_LEFT)
			}
		} else if g.LeftFighter.Timer.Value < g.RightFighter.Timer.Value {
			g.Act(LEFT_TO_RIGHT)
		} else { // g.RightFighter.Timer < g.LeftFighter.Timer
			g.Act(RIGHT_TO_LEFT)
		}
	} else if lReady {
		g.Act(LEFT_TO_RIGHT)
	} else if rReady {
		g.Act(RIGHT_TO_LEFT)
	}
	g.LeftFighter.Timer.Value -= g.LeftFighter.Speed.Value
	g.RightFighter.Timer.Value -= g.RightFighter.Speed.Value
}
