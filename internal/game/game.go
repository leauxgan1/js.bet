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
	fighter.FighterAnim = "attack"
	if !hit {
		g.AudioPlayers.DodgePlaying = true
		eventlog.EventLog.Write(fmt.Sprintf("%s just missed...", fighter.Name))
		return
	} else {
		g.AudioPlayers.AttackPlaying = true
	}
	oppFighter.FighterAnim = "defend"
	g.AudioPlayers.BlockPlaying = true
	crit := fighter.CheckCrit()
	if crit {
		g.AudioPlayers.AttackPlaying = false
		g.AudioPlayers.CritPlaying = true
		damage *= 2.0
		fighter.FighterAnim = "crit"
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

func useAbility(abilityIdx int, self *Fighter, other *Fighter) {
	ability := self.Abilities[abilityIdx]
	ability.InvokeFunc(self, other)
	eventlog.EventLog.Write(fmt.Sprintf("%s used '%s' ", self.Name, ability.Name))
	self.Abilities[abilityIdx].Timer.Value = ability.Timer.MaxValue
	self.FighterAnim = "ability"
}

func (g *GameState) StepGame() {
	g.FrameCount += 1
	g.RightFighter.FighterAnim = "idle"
	g.LeftFighter.FighterAnim = "idle"
	g.AudioPlayers.Stop()

	// Check for a non-positive health, choose a winner and keep them in the game for the next round
	var winner = determineWinner(g.LeftFighter, g.RightFighter)
	if winner == LEFT || winner == RIGHT {
		g.Winner = winner
		g.ResetKeepWinner()
		return
	}

	// Check if an ability is ready on each fighter to see if they should use it, prioritize ability usage over attacks
	var leftAbilityIdx int = -1
	var rightAbilityIdx int = -1
	// Update all ability timers on each fighter
	for i := 0; i < len(g.LeftFighter.Abilities); i++ {
		ability := g.LeftFighter.Abilities[i]
		g.LeftFighter.Abilities[i].Timer.Value -= 1
		if ability.Timer.Value <= 0 {
			leftAbilityIdx = i
		}
	}
	for i := 0; i < len(g.RightFighter.Abilities); i++ {
		ability := g.RightFighter.Abilities[i]
		g.RightFighter.Abilities[i].Timer.Value -= 1
		if ability.Timer.Value <= 0 {
			rightAbilityIdx = i
		}
	}

	// Update all effect durations on each fighter
	for i, effect := range g.LeftFighter.Effects {
		// Reduce effect duration if > 0
		if effect.GetDuration() > 0 {
			effect.StepDuration()
		} else {
			g.LeftFighter.Effects = slices.Delete(g.LeftFighter.Effects, i, i+1)
		}
		// Apply tick function on each fighter
		effect.OnTick(&g.LeftFighter)
	}
	for i, effect := range g.RightFighter.Effects {
		// Reduce effect duration if > 0
		if effect.GetDuration() > 0 {
			effect.StepDuration()
		} else {
			g.RightFighter.Effects = slices.Delete(g.RightFighter.Effects, i, i+1)
		}
		// Apply tick function on each fighter
		effect.OnTick(&g.RightFighter)
	}

	// Prioritize using an ability first and then exiting
	if leftAbilityIdx != -1 || rightAbilityIdx != -1 {
		if leftAbilityIdx != -1 && rightAbilityIdx != -1 {
			leftAbility := g.LeftFighter.Abilities[leftAbilityIdx]
			rightAbility := g.RightFighter.Abilities[rightAbilityIdx]
			if leftAbility.Timer.Value < rightAbility.Timer.Value {
				useAbility(leftAbilityIdx, &g.LeftFighter, &g.RightFighter)
			} else if leftAbility.Timer.Value > rightAbility.Timer.Value {
				useAbility(rightAbilityIdx, &g.RightFighter, &g.LeftFighter)
			} else {
				rand := rand.Float32() // Choose randomly on second tie
				if rand < 0.5 {        // Left fighter acts
					useAbility(leftAbilityIdx, &g.LeftFighter, &g.RightFighter)
				} else { // Right fighter acts
					useAbility(rightAbilityIdx, &g.RightFighter, &g.LeftFighter)
				}
			}
		} else if leftAbilityIdx != -1 {
			useAbility(leftAbilityIdx, &g.LeftFighter, &g.RightFighter)
		} else { // Right ability is not nil
			useAbility(rightAbilityIdx, &g.RightFighter, &g.LeftFighter)
		}
		return // Return regardless to not double dip
	}

	// Step forward each fighter's attack timer
	g.LeftFighter.AttackTimer.Value -= g.LeftFighter.Speed.Value
	g.RightFighter.AttackTimer.Value -= g.RightFighter.Speed.Value

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
}
