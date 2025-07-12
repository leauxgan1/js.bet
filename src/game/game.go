package game

import (
	"code-root/src/eventlog"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
)

var bets map[string]int = make(map[string]int)
func setBet(name string, amount int) {
	bets[name] = amount
}

func clearBets() {
	for k := range bets {
		delete(bets, k)
	}
}

type Action struct {
	Direction InitiativeEnum
	WasHit bool
	WasCrit bool
}

type GameState struct {
	LeftFighter Fighter
	RightFighter Fighter
	Winner WinnerEnum
	FrameCount int
	LastAction Action
	AudioPlayers AudioPlayer
}

func New() GameState {
	leftFighter := chooseRandomFighter()
	rightFighter,err := chooseRandomFighterExclusive(leftFighter.Name)
	if err != nil {
		log.Panic(err)
	}
	return GameState {
		LeftFighter: leftFighter,
		RightFighter: rightFighter,
		Winner: NEITHER,
	}
}

func (g *GameState) ResetKeepWinner() {
	if g.Winner == LEFT {
		g.LeftFighter.Reset()
		newRight, err := chooseRandomFighterExclusive(g.LeftFighter.Name)
		if err != nil {
			return 
		}
		g.RightFighter = newRight
	} else if g.Winner == RIGHT {
		g.RightFighter.Reset()
		newLeft, err := chooseRandomFighterExclusive(g.RightFighter.Name)
		if err != nil {
			return 
		}
		g.LeftFighter = newLeft
	} else {
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

type FighterState uint 
const (
	_ = iota
	READY
	DEFENDING
	ATTACKING
	CRITTING
	DYING
)

type Fighter struct {
	Name string         // Name of framework/library
	Health int          // Represents how much of a "industry standard" the framework/library is / likelihood to stick around or be popular
	MaxHealth int       
	Damage int          // Represents how consistently useful the framework/library is for common tasks
	Speed int           // Represents the overall performance under load and scalability of the framework/library, causes fighter to act sooner
	Timer int				    // Time before next action of fighter, reduced by speed each turn
	Accuracy float32    // Represents how simple the library/frame work is / how easy it is to get it right at first, causes less misses
	CritRate float32    // Represents how suprisingly useful or versatile the framework/library is in niche situations
	State FighterState
}
const DEFAULT_TIMER = 25

func (f *Fighter) Reset() *Fighter {
	f.State = READY
	f.Health = f.MaxHealth
	return f
}

var fighterList = [...]Fighter {
	{
		Name: "JQuery",
		Health: 30,
		MaxHealth: 30,
		Damage: 4,
		Speed: 8,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.5,
		CritRate: 0.0,
	},
	{
		Name: "React",
		Health: 20,
		MaxHealth: 20,
		Damage: 5,
		Speed: 3,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.6,
		CritRate: 0.2,
	},
	{
		Name: "Svelte",
		Health: 16,
		MaxHealth: 16,
		Damage: 5,
		Speed: 7,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.8,
		CritRate: 0.4,
	},
	{
		Name: "Solid",
		Health: 16,
		MaxHealth: 16,
		Damage: 6,
		Speed: 7,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.8,
		CritRate: 0.3,
	},
	{
		Name: "HTMX",
		Health: 10,
		MaxHealth: 10,
		Damage: 10,
		Speed: 8,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.99,
		CritRate: 0.4,
	},
	{
		Name: "Datastar",
		Health: 8,
		MaxHealth: 8,
		Damage: 11,
		Speed: 9,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.99,
		CritRate: 0.4,
	},
}

func chooseRandomFighter() Fighter {
	randomIndex := rand.IntN(len(fighterList))
	randomFighter := fighterList[randomIndex]
	log.Printf("Randomly chose %v\n",randomFighter)
	return randomFighter
}

func chooseRandomFighterExclusive(excludedFighterName string) (Fighter, error) {
	swapIndex := -1
	for i := 0; i < len(fighterList); i += 1 {
		if (fighterList[i].Name == excludedFighterName) {
			swapIndex = i
			break
		}
	}
	if swapIndex == -1 {
		return Fighter{}, errors.New(fmt.Sprintf("Error: Fighter name %s not found",excludedFighterName))
	}
	// Swap excluded fighter with first index
	temp := fighterList[0]
	fighterList[0] = fighterList[swapIndex]
	fighterList[swapIndex] = temp

	randomIndex := 1 + rand.IntN(len(fighterList) - 1) // Choose from [1,1-len)
	randomFighter := fighterList[randomIndex]
	log.Printf("Randomly chose %v, excluding %s\n",randomFighter,excludedFighterName)
	return randomFighter, nil
}

type InitiativeEnum bool
const (
	LEFT_TO_RIGHT = true
	RIGHT_TO_LEFT = false
)

func (g *GameState) Act(initiative InitiativeEnum)  {
	var crit bool = false
	if initiative == LEFT_TO_RIGHT {
		g.LeftFighter.Timer = DEFAULT_TIMER

		hit := g.LeftFighter.CheckHit()
		damage := g.LeftFighter.Damage
		g.LeftFighter.State = ATTACKING
		if !hit {
			g.AudioPlayers.DodgePlaying = true
			eventlog.EventLog.Write(fmt.Sprintf("%s just missed...",g.LeftFighter.Name))
			return
		} else {
			g.AudioPlayers.AttackPlaying = true
		}
		g.RightFighter.State = DEFENDING
		g.AudioPlayers.BlockPlaying = true
		crit = g.LeftFighter.CheckCrit()
		if crit {
			g.AudioPlayers.AttackPlaying = false
			g.AudioPlayers.CritPlaying = true
			damage *= 2.0
			g.LeftFighter.State = CRITTING
			eventlog.EventLog.Write(fmt.Sprintf("%s just crit %s for %d",g.LeftFighter.Name,g.RightFighter.Name,damage))
		} else {
			eventlog.EventLog.Write(fmt.Sprintf("%s just hit %s for %d",g.LeftFighter.Name,g.RightFighter.Name,damage))
		} 		
		g.RightFighter.Health -= damage
	} else { // RIGHT_TO_LEFT
		g.RightFighter.Timer = DEFAULT_TIMER

		hit := g.RightFighter.CheckHit()
		damage := g.LeftFighter.Damage
		g.RightFighter.State = ATTACKING
		if !hit {
			g.AudioPlayers.DodgePlaying = true
			eventlog.EventLog.Write(fmt.Sprintf("%s just missed...",g.RightFighter.Name))
			return
		} else {
			g.AudioPlayers.AttackPlaying = true
		}

		g.LeftFighter.State = DEFENDING
		g.AudioPlayers.BlockPlaying = true
		crit = g.LeftFighter.CheckCrit()
		if crit {
			g.AudioPlayers.AttackPlaying = false
			g.AudioPlayers.CritPlaying = true
			damage *= 2.0
			g.RightFighter.State = CRITTING
			eventlog.EventLog.Write(fmt.Sprintf("%s just crit %s for %d",g.RightFighter.Name,g.LeftFighter.Name,damage))
		} else {
			eventlog.EventLog.Write(fmt.Sprintf("%s just hit %s for %d",g.RightFighter.Name,g.LeftFighter.Name,damage))
		}
		g.LeftFighter.Health -= damage
	}
}
func (f Fighter) CheckHit() bool {
	if f.Accuracy > 0.0 && rand.Float32() < f.Accuracy {
		return true
	}
	return false
}

func (f Fighter) CheckCrit() bool {
	if f.CritRate > 0.0 && rand.Float32() < f.CritRate {
		return true
	}
	return false
}

// TODO add sound for winner being determined
func determineWinner(left Fighter, right Fighter) WinnerEnum {
	if(left.Health <= 0 && right.Health <= 0) {
		if left.Health < right.Health {
			return RIGHT
		} else {
			return LEFT
		}
	} else if (left.Health <= 0) {
		return RIGHT
	} else if (right.Health <= 0) {
		return LEFT
	}
	return NEITHER
}

func (g *GameState) StepGame()  {
	// log.Printf("Game Running on frame %d:\n",g.FrameCount)
	g.FrameCount += 1
	g.RightFighter.State = READY
	g.LeftFighter.State = READY
	g.AudioPlayers.Stop()
	// Check for a non-positive health, choose a winner and keep them in the game for the next round
	var winner WinnerEnum = determineWinner(g.LeftFighter,g.RightFighter)
	if winner == LEFT || winner == RIGHT {
		g.Winner = winner
		g.ResetKeepWinner()
		return
	}

	lReady := g.LeftFighter.Timer <= 0
	rReady := g.RightFighter.Timer <= 0
	if lReady && rReady {
		if g.LeftFighter.Timer == g.RightFighter.Timer { // Choose lesser Timer when both ready, higher speed on ties
			if g.LeftFighter.Speed == g.RightFighter.Speed {
				rand := rand.Float32() // Choose randomly on second tie
				if rand < 0.5 { // Left fighter acts
					g.Act(LEFT_TO_RIGHT)
				} else {        // Right fighter acts
					g.Act(RIGHT_TO_LEFT)
				}
			} else if g.LeftFighter.Speed > g.RightFighter.Speed { 
				g.Act(LEFT_TO_RIGHT)
			} else {
				g.Act(RIGHT_TO_LEFT)
			}
		} else if g.LeftFighter.Timer < g.RightFighter.Timer {
				g.Act(LEFT_TO_RIGHT)
		} else { // g.RightFighter.Timer < g.LeftFighter.Timer
				g.Act(RIGHT_TO_LEFT)
		}
	} else if lReady {
		g.Act(LEFT_TO_RIGHT)
	} else if rReady {
		g.Act(RIGHT_TO_LEFT)
	} 
	g.LeftFighter.Timer -= g.LeftFighter.Speed
	g.RightFighter.Timer -= g.RightFighter.Speed
}
