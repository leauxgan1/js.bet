package game

import (
	"fmt"
	"js-bet/internal/eventlog"
	"log"
	"math/rand/v2"
	"slices"
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

	fighter.AttackTimer.Value = fighter.AttackTimer.MaxValue // Reset timer

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

	// Check if an ability is ready on each fighter to see if they should use it
	// > Prioritize first found abiity in list
	// > Prioritize fighter with lower cooldown
	var leftFighterAbility *Ability
	var rightFighterAbility *Ability

	for _, ability := range g.LeftFighter.Abilities {
		if ability.Timer.Value <= 0 {
			leftFighterAbility = &ability
		}
	}
	for _, ability := range g.RightFighter.Abilities {
		if ability.Timer.Value <= 0 {
			rightFighterAbility = &ability
		}
	}
	if leftFighterAbility != nil {
		if rightFighterAbility != nil {
			if leftFighterAbility.Timer.Value > rightFighterAbility.Timer.Value {
				rightFighterAbility.InvokeFunc(&g.RightFighter, &g.LeftFighter)
			} else if leftFighterAbility.Timer.Value < rightFighterAbility.Timer.Value {
				leftFighterAbility.InvokeFunc(&g.LeftFighter, &g.RightFighter)
			} else {
				rand := rand.Float32() // Choose randomly on second tie
				if rand < 0.5 {        // Left fighter acts
					leftFighterAbility.InvokeFunc(&g.LeftFighter, &g.RightFighter)
				} else { // Right fighter acts
					rightFighterAbility.InvokeFunc(&g.RightFighter, &g.LeftFighter)
				}
			}
			return
		}
		leftFighterAbility.InvokeFunc(&g.LeftFighter, &g.RightFighter)
		return
	}
	if rightFighterAbility != nil {
		rightFighterAbility.InvokeFunc(&g.RightFighter, &g.LeftFighter)
		return
	}

	// Todo: Skip doing attack when ability was used

	lReady := g.LeftFighter.AttackTimer.Value <= 0
	rReady := g.RightFighter.AttackTimer.Value <= 0

	if lReady && rReady {
		if g.LeftFighter.AttackTimer.Value == g.RightFighter.AttackTimer.Value { // Choose lesser AttackTimer when both ready, higher speed on ties
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
		} else if g.LeftFighter.AttackTimer.Value < g.RightFighter.AttackTimer.Value {
			g.Act(LEFT_TO_RIGHT)
		} else { // g.RightFighter.AttackTimer < g.LeftFighter.AttackTimer
			g.Act(RIGHT_TO_LEFT)
		}
	} else if lReady {
		g.Act(LEFT_TO_RIGHT)
	} else if rReady {
		g.Act(RIGHT_TO_LEFT)
	}

	g.LeftFighter.AttackTimer.Value -= g.LeftFighter.Speed.Value
	g.RightFighter.AttackTimer.Value -= g.RightFighter.Speed.Value

	// Update all ability timers on each fighter
	for _, ability := range g.LeftFighter.Abilities {
		ability.Timer.Value -= 0
	}
	for _, ability := range g.RightFighter.Abilities {
		ability.Timer.Value -= 0
	}

	// Update all effect durations on each fighter
	for i, effect := range g.LeftFighter.Effects {
		// Reduce effect duration if > 0
		if effect.duration.Value > 0 {
			effect.duration.Value -= 1
		} else {
			g.LeftFighter.Effects = slices.Delete(g.LeftFighter.Effects, i, i+1)
		}
		// Apply tick function on each fighter
		effect.tickFunc(&g.LeftFighter)
	}
	for i, effect := range g.RightFighter.Effects {
		// Reduce effect duration if > 0
		if effect.duration.Value > 0 {
			effect.duration.Value -= 1
		} else {
			g.RightFighter.Effects = slices.Delete(g.RightFighter.Effects, i, i+1)
		}
		// Apply tick function on each fighter
		effect.tickFunc(&g.RightFighter)
	}

}
