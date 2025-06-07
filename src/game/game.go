package game

import (
	"code-root/src/eventlog"
	"errors"
	"fmt"
	"log"
	"math/rand/v2"
)

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
)

type Fighter struct {
	Name string         // Name of framework/library
	Health int          // Represents how much of a "industry standard" the framework/library is / likelihood to stick around or be popular
	maxHealth int       
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
	f.Health = f.maxHealth
	return f
}

var fighterList = [...]Fighter {
	{
		Name: "JQuery",
		Health: 15,
		maxHealth: 15,
		Damage: 4,
		Speed: 8,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.4,
		CritRate: 0.0,
	},
	{
		Name: "React",
		Health: 10,
		maxHealth: 10,
		Damage: 5,
		Speed: 3,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.5,
		CritRate: 0.2,
	},
	{
		Name: "Svelte",
		Health: 8,
		maxHealth: 8,
		Damage: 5,
		Speed: 7,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.7,
		CritRate: 0.4,
	},
	{
		Name: "Solid",
		Health: 8,
		maxHealth: 8,
		Damage: 6,
		Speed: 7,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.7,
		CritRate: 0.3,
	},
	{
		Name: "HTMX",
		Health: 5,
		maxHealth: 5,
		Damage: 10,
		Speed: 8,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.9,
		CritRate: 0.4,
	},
	{
		Name: "Datastar",
		Health: 4,
		maxHealth: 4,
		Damage: 11,
		Speed: 9,
		Timer: DEFAULT_TIMER,
		Accuracy: 0.9,
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
	var crit bool
	if initiative == LEFT_TO_RIGHT {
		g.LeftFighter.Timer = DEFAULT_TIMER
		
		hit := g.LeftFighter.CheckHit()
		damage := g.LeftFighter.Damage
		g.LeftFighter.State = ATTACKING
		if !hit {
			eventlog.EventLog.Write(fmt.Sprintf("%s just missed...",g.LeftFighter.Name))
			return
		}
		g.RightFighter.State = DEFENDING
		crit = g.LeftFighter.CheckCrit()
		if crit {
			damage *= 2.0
			g.LeftFighter.State = CRITTING
			eventlog.EventLog.Write(fmt.Sprintf("%s just crit %s for %d",g.LeftFighter.Name,g.RightFighter.Name,damage))
		} 		
		g.RightFighter.Health -= damage
	} else { // RIGHT_TO_LEFT
		g.RightFighter.Timer = DEFAULT_TIMER

		hit := g.RightFighter.CheckHit()
		damage := g.LeftFighter.Damage
		g.RightFighter.State = ATTACKING
		if !hit {
			eventlog.EventLog.Write(fmt.Sprintf("%s just missed...",g.RightFighter.Name))
			return
		}
		g.LeftFighter.State = DEFENDING
		crit = g.LeftFighter.CheckCrit()
		if crit {
			damage *= 2.0
			g.RightFighter.State = CRITTING
			eventlog.EventLog.Write(fmt.Sprintf("%s just crit %s for %d",g.RightFighter.Name,g.LeftFighter.Name,damage))
		} 		
		g.LeftFighter.Health -= damage
	}
}
func (f Fighter) CheckHit() bool {
	if f.Accuracy > 0.0 && rand.Float32() < f.CritRate {
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

func (g *GameState) StepGame()  {
	log.Printf("Game Running on frame %d with log:\n",g.FrameCount)
	g.FrameCount += 1
	g.RightFighter.State = READY
	g.LeftFighter.State = READY

	lReady := g.LeftFighter.Timer <= 0
	rReady := g.RightFighter.Timer <= 0
	if lReady && rReady {
		// Choose lesser Timer when both ready, higher speed on ties
		if g.LeftFighter.Timer == g.RightFighter.Timer {
			// Choose randomly on second tie
			if g.LeftFighter.Speed == g.RightFighter.Speed {
				rand := rand.Float32()
				if rand < 0.5 { // Left fighter acts
					g.Act(LEFT_TO_RIGHT)
				} else { // Right fighter acts
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
	// Upon getting a non-positive health, choose a winner and keep them in the game for the next round
	if(g.LeftFighter.Health <= 0 && g.RightFighter.Health <= 0) {
		if g.LeftFighter.Health == g.RightFighter.Health {
			g.Winner = NEITHER
		} else if g.LeftFighter.Health < g.RightFighter.Health {
			g.Winner = RIGHT
		} else {
			g.Winner = LEFT
		}
		g.ResetKeepWinner()
		return
	} else if (g.LeftFighter.Health <= 0) {
		g.Winner = RIGHT
		g.ResetKeepWinner()
		return
	} else if (g.RightFighter.Health <= 0) {
		g.Winner = LEFT
		g.ResetKeepWinner()
		return
	}
	g.LeftFighter.Timer -= g.LeftFighter.Speed
	g.RightFighter.Timer -= g.RightFighter.Speed
}
