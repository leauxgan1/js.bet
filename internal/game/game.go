package game

import (
	"fmt"
	"js-bet/internal/eventlog"
	"log"
	"math/rand/v2"
	// "slices"
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

type UserState struct {
	Gold int
}

func New() GameState {

	fighters := [2]Fighter{}
	fighters[0] = chooseReact()
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

func (g *GameState) Act(order ActingOrder) {
	fighterIdx := 0
	oppIdx := 1
	switch order {
	case RIGHTTOLEFT:
		fighterIdx = 1
		oppIdx = 0
	}
	// Reset actor's attack timer to its maximum
	g.Fighters[fighterIdx].AttackTimer.Value = g.Fighters[fighterIdx].AttackTimer.MaxValue // Reset timer
	// Determine if hit was confirmed
	hit := g.Fighters[fighterIdx].CheckHit(0.0)
	damage := g.Fighters[fighterIdx].Damage.Value
	g.Fighters[fighterIdx].FighterAnim = "attack"
	if !hit {
		g.AudioPlayers.DodgePlaying = true
		eventlog.EventLog.Write(fmt.Sprintf("%s just missed...", g.Fighters[fighterIdx].Name))
		// Add message to the status bar
		g.Status = fmt.Sprintf("%s attacked %s and missed!", g.Fighters[fighterIdx].Name, g.Fighters[oppIdx].Name)
		return
	} else {
		g.AudioPlayers.AttackPlaying = true
		// Add message to the status bar
		g.Status = fmt.Sprintf("%s attacked %s", g.Fighters[fighterIdx].Name, g.Fighters[oppIdx].Name)
	}
	g.Fighters[oppIdx].FighterAnim = "defend"
	g.AudioPlayers.BlockPlaying = true
	crit := g.Fighters[fighterIdx].CheckCrit()
	if crit {
		g.AudioPlayers.AttackPlaying = false
		g.AudioPlayers.CritPlaying = true
		damage *= 2.0
		g.Fighters[fighterIdx].FighterAnim = "crit"
		eventlog.EventLog.Write(fmt.Sprintf("%s just critically hit %s for %d", g.Fighters[fighterIdx].Name, g.Fighters[oppIdx].Name, damage))
	} else {
		eventlog.EventLog.Write(fmt.Sprintf("%s just hit %s for %d", g.Fighters[fighterIdx].Name, g.Fighters[oppIdx].Name, damage))
	}
	g.Fighters[oppIdx].Health.Value -= damage
}

func (f Fighter) CheckHit(dodgeRate float32) bool {
	if f.Accuracy.Value > 0.0 && rand.Float32() < f.Accuracy.Value-dodgeRate {
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
func (g *GameState) determineWinner() WinnerEnum {
	left := g.Fighters[0]
	right := g.Fighters[1]
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
	var winner = g.determineWinner()
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
	usedAbilityIdxs := [2]int{-1, -1}

	// For each fighter...
	for fIdx := 0; fIdx < 2; fIdx += 1 {
		// Update all effect durations on each fighter
		log.Printf("%d -> %v", fIdx, g.Fighters[fIdx].Effects)
		for i, effect := range g.Fighters[fIdx].Effects {
			// Reduce effect duration if > 0
			if effect.GetDuration() > 0 {
				effect.StepDuration()
			} else {
				g.Fighters[fIdx].Effects = append(g.Fighters[fIdx].Effects[:i], g.Fighters[fIdx].Effects[i+1:]...)
			}
			// Apply tick function on each fighter
			effect.OnTick(&g.Fighters[0])
		}
		// Update all ability timers on each fighter
		for i := 0; i < len(g.Fighters[fIdx].Abilities); i++ {
			ability := g.Fighters[fIdx].Abilities[i]
			g.Fighters[fIdx].Abilities[i].Timer.Value -= 1
			if ability.Timer.Value <= 0 {
				usedAbilityIdxs[fIdx] = i
			}
		}
	}
	var leftAbilityIdx int = usedAbilityIdxs[0]
	var rightAbilityIdx int = usedAbilityIdxs[1]

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
		return // Return regardless to not double dip on abilities and attacks
	}

	// Step forward each fighter's attack timer
	g.Fighters[0].AttackTimer.Value -= g.Fighters[0].Speed.Value
	g.Fighters[1].AttackTimer.Value -= g.Fighters[1].Speed.Value

	order := g.determineActingOrder()
	switch order {
	case NOT_READY:
		break
	default:
		g.Act(order)
	}
}

type ActingOrder uint

const (
	_ = iota
	LEFTTORIGHT
	RIGHTTOLEFT
	NOT_READY
)

func (g *GameState) determineActingOrder() ActingOrder {
	lReady := g.Fighters[0].AttackTimer.Value <= 0
	rReady := g.Fighters[1].AttackTimer.Value <= 0

	if lReady && rReady {
		if g.Fighters[0].AttackTimer.Value == g.Fighters[1].AttackTimer.Value { // Choose lesser AttackTimer when both ready, higher speed on ties
			if g.Fighters[0].Speed.Value == g.Fighters[1].Speed.Value {
				rand := rand.Float32() // Choose randomly on second tie
				if rand < 0.5 {        // Left fighter acts
					return LEFTTORIGHT
				} else { // Right fighter acts
					return RIGHTTOLEFT
				}
			} else if g.Fighters[0].Speed.Value > g.Fighters[1].Speed.Value {
				return LEFTTORIGHT
			} else {
				return RIGHTTOLEFT
			}
		} else if g.Fighters[0].AttackTimer.Value < g.Fighters[1].AttackTimer.Value {
			return LEFTTORIGHT
		} else { // g.Fighters[1].AttackTimer < g.Fighters[0].AttackTimer
			return RIGHTTOLEFT
		}
	} else if lReady {
		return LEFTTORIGHT
	} else if rReady {
		return RIGHTTOLEFT
	}
	return NOT_READY // Reduntant return
}
