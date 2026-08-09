package game

import (
	"fmt"
	"js-bet/internal/eventlog"
	"log"
	"math/rand/v2"
	"slices"
)

// Statistical Consts
const CRIT_MULTIPLIER = 2.0

type GameState struct {
	Fighters     [2]Fighter
	FrameCount   int
	AudioPlayers AudioPlayer
	Winner       WinnerEnum
	Phase        GamePhase
	PhaseTimer   int // Timer for pre-round and post-round phases (Not an IntValue since each has its own duration)
	Status       string
}

func New() GameState {

	fighters := [2]Fighter{}
	fighters[0] = chooseRandomFighter()
	rightFighter, err := chooseRandomFighterExclusive(fighters[0].Name)
	if err != nil {
		log.Panic(err)
	}
	fighters[1] = rightFighter
	return GameState{
		Fighters:   fighters,
		Winner:     NEITHER,
		Phase:      PREROUND,
		PhaseTimer: 10,
	}
}

func (g *GameState) ResetKeepWinner() {
	switch g.Winner {
	case LEFT:
		g.Fighters[0].Reset()
		newRight, err := chooseRandomFighterExclusive(g.Fighters[0].Name)
		if err != nil {
			return
		}
		g.Fighters[1] = newRight
	case RIGHT:
		g.Fighters[1].Reset()
		newLeft, err := chooseRandomFighterExclusive(g.Fighters[1].Name)
		if err != nil {
			return
		}
		g.Fighters[0] = newLeft
	default:
		*g = New()
	}
}

type WinnerEnum uint

const (
	_       = iota
	LEFT    // 1
	RIGHT   // 2
	NEITHER // 3
)

type GamePhase uint

const (
	_ = iota
	PREROUND
	ROUND
	POSTROUND
)

func (g *GameState) Act(fighter *Fighter, oppFighter *Fighter) {
	// Reset actor's attack timer to its maximum
	fighter.AttackTimer.Value = fighter.AttackTimer.MaxValue // Reset timer
	// Determine if hit was confirmed
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

	// Add message to the status bar
	g.Status = fmt.Sprintf("%s attacked %s", fighter.Name, oppFighter.Name)
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

func useAbility(abilityIdx int, self *Fighter, other *Fighter, gs *GameState) {
	ability := self.Abilities[abilityIdx]
	ability.InvokeFunc(self, other)
	eventlog.EventLog.Write(fmt.Sprintf("%s used '%s' ", self.Name, ability.Name))
	self.Abilities[abilityIdx].Timer.Value = ability.Timer.MaxValue
	self.FighterAnim = "ability"
	gs.Status = fmt.Sprintf("%s used '%s'", self.Name, ability.Name)
}

func (g *GameState) StepGame() {
	switch g.Phase {
	case PREROUND:
		g.PhaseTimer -= 1
		g.Status = "Pre-Round Phase"
		if g.PhaseTimer < 0 {
			g.Phase = ROUND
			g.Status = "Round start!"
		}
		return
	case POSTROUND:
		g.PhaseTimer -= 1
		if g.PhaseTimer < 0 {
			g.Phase = PREROUND
			g.ResetKeepWinner()
			g.PhaseTimer = 10
			g.Status = "Pre-Round Phase"
			g.Winner = NEITHER
		}
		return
	default: // Skip if middle of round or any other case
		break
	}

	g.FrameCount += 1
	g.Fighters[0].FighterAnim = "idle"
	g.Fighters[1].FighterAnim = "idle"
	g.AudioPlayers.Stop()

	// Check for a non-positive health, choose a winner and keep them in the game for the next round
	var winner = determineWinner(g.Fighters[0], g.Fighters[1])
	if winner != NEITHER {
		g.Winner = winner
		g.Phase = POSTROUND
		g.PhaseTimer = 10
		var winnerName string
		switch g.Winner {
		case LEFT:
			winnerName = g.Fighters[0].Name
		case RIGHT:
			winnerName = g.Fighters[1].Name
		}
		g.Status = fmt.Sprintf("Winner is: %s", winnerName)
		return
	}

	// Check if an ability is ready on each fighter to see if they should use it, prioritize ability usage over attacks
	var leftAbilityIdx int = -1
	var rightAbilityIdx int = -1

	// Update all ability timers on each fighter
	for i := 0; i < len(g.Fighters[0].Abilities); i++ {
		ability := g.Fighters[0].Abilities[i]
		g.Fighters[0].Abilities[i].Timer.Value -= 1
		if ability.Timer.Value <= 0 {
			leftAbilityIdx = i
		}
	}
	for i := 0; i < len(g.Fighters[1].Abilities); i++ {
		ability := g.Fighters[1].Abilities[i]
		g.Fighters[1].Abilities[i].Timer.Value -= 1
		if ability.Timer.Value <= 0 {
			rightAbilityIdx = i
		}
	}

	// Update all effect durations on each fighter
	for i, effect := range g.Fighters[0].Effects {
		// Reduce effect duration if > 0
		if effect.GetDuration() > 0 {
			effect.StepDuration()
		} else {
			g.Fighters[0].Effects = slices.Delete(g.Fighters[0].Effects, i, i+1)
		}
		// Apply tick function on each fighter
		effect.OnTick(&g.Fighters[0])
	}
	for i, effect := range g.Fighters[1].Effects {
		// Reduce effect duration if > 0
		if effect.GetDuration() > 0 {
			effect.StepDuration()
		} else {
			g.Fighters[1].Effects = slices.Delete(g.Fighters[1].Effects, i, i+1)
		}
		// Apply tick function on each fighter
		effect.OnTick(&g.Fighters[1])
	}

	// Prioritize using an ability first and then exiting
	if leftAbilityIdx != -1 || rightAbilityIdx != -1 {
		if leftAbilityIdx != -1 && rightAbilityIdx != -1 {
			leftAbility := g.Fighters[0].Abilities[leftAbilityIdx]
			rightAbility := g.Fighters[1].Abilities[rightAbilityIdx]
			if leftAbility.Timer.Value < rightAbility.Timer.Value {
				useAbility(leftAbilityIdx, &g.Fighters[0], &g.Fighters[1], g)
			} else if leftAbility.Timer.Value > rightAbility.Timer.Value {
				useAbility(rightAbilityIdx, &g.Fighters[1], &g.Fighters[0], g)
			} else {
				rand := rand.Float32() // Choose randomly on second tie
				if rand < 0.5 {        // Left fighter acts
					useAbility(leftAbilityIdx, &g.Fighters[0], &g.Fighters[1], g)
				} else { // Right fighter acts
					useAbility(rightAbilityIdx, &g.Fighters[1], &g.Fighters[0], g)
				}
			}
		} else if leftAbilityIdx != -1 {
			useAbility(leftAbilityIdx, &g.Fighters[0], &g.Fighters[1], g)
		} else { // Right ability is not nil
			useAbility(rightAbilityIdx, &g.Fighters[1], &g.Fighters[0], g)
		}
		return // Return regardless to not double dip
	}

	// Step forward each fighter's attack timer
	g.Fighters[0].AttackTimer.Value -= g.Fighters[0].Speed.Value
	g.Fighters[1].AttackTimer.Value -= g.Fighters[1].Speed.Value

	// Todo: Skip doing attack when ability was used
	lReady := g.Fighters[0].AttackTimer.Value <= 0
	rReady := g.Fighters[1].AttackTimer.Value <= 0

	if lReady && rReady {
		if g.Fighters[0].AttackTimer.Value == g.Fighters[1].AttackTimer.Value { // Choose lesser AttackTimer when both ready, higher speed on ties
			if g.Fighters[0].Speed.Value == g.Fighters[1].Speed.Value {
				rand := rand.Float32() // Choose randomly on second tie
				if rand < 0.5 {        // Left fighter acts
					g.Act(&g.Fighters[0], &g.Fighters[1])
				} else { // Right fighter acts
					g.Act(&g.Fighters[1], &g.Fighters[0])
				}
			} else if g.Fighters[0].Speed.Value > g.Fighters[1].Speed.Value {
				g.Act(&g.Fighters[0], &g.Fighters[1])
			} else {
				g.Act(&g.Fighters[1], &g.Fighters[0])
			}
		} else if g.Fighters[0].AttackTimer.Value < g.Fighters[1].AttackTimer.Value {
			g.Act(&g.Fighters[0], &g.Fighters[1])
		} else { // g.Fighters[1].AttackTimer < g.Fighters[0].AttackTimer
			g.Act(&g.Fighters[1], &g.Fighters[0])
		}
	} else if lReady {
		g.Act(&g.Fighters[0], &g.Fighters[1])
	} else if rReady {
		g.Act(&g.Fighters[1], &g.Fighters[0])
	}
}

func (g *GameState) determineActingOrder() {

}
