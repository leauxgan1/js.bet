package game

import (
	"fmt"
	"js-bet/internal/eventlog"
	"log"
	"math/rand/v2"
)

var Bets map[string]int = make(map[string]int)

func SetBet(name string, amount int) {
	Bets[name] = amount
}

func ClearBets() {
	for k := range Bets {
		delete(Bets, k)
	}
}

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

type FighterState uint

const (
	_ = iota
	READY
	DEFENDING
	ATTACKING
	CRITTING
	DYING
)

type IntStat struct {
	Value    int
	MaxValue int
}

func NewIntStat(value int) IntStat {
	return IntStat{
		Value:    value,
		MaxValue: value,
	}
}

type FloatStat struct {
	Value    float32
	MaxValue float32
}

func NewFloatStat(value float32) FloatStat {
	return FloatStat{
		Value:    value,
		MaxValue: value,
	}
}

type Fighter struct {
	Name     string    // Name of framework/library
	Color    string    // Color of logo
	Health   IntStat   // Represents how much of an "industry standard" the framework/library is / likelihood to stick around in the future
	Damage   IntStat   // Represents how consistently useful the framework/library is for common tasks
	Speed    IntStat   // Represents the overall performance under load and scalability of the framework/library, causes fighter to act sooner
	Timer    IntStat   // Time before next action of fighter, reduced by speed each turn
	Accuracy FloatStat // Represents how simple the library/frame work is / how easy it is to get it right at first, causes less misses
	CritRate FloatStat // Represents how suprisingly useful or versatile the framework/library is in niche situations
	State    FighterState
}

func (f *Fighter) Reset() *Fighter {
	f.State = READY
	f.Health.Value = f.Health.MaxValue
	return f
}

var fighterList = [...]Fighter{
	{
		Name:     "JQuery",
		Color:    "#0769AD",
		Health:   NewIntStat(30),
		Damage:   NewIntStat(4),
		Speed:    NewIntStat(8),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.5),
		CritRate: NewFloatStat(0.0),
	},
	{
		Name:     "React",
		Color:    "#58C4DC",
		Health:   NewIntStat(20),
		Damage:   NewIntStat(5),
		Speed:    NewIntStat(3),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.6),
		CritRate: NewFloatStat(0.2),
	},
	{
		Name:     "Svelte",
		Color:    "#FF5018",
		Health:   NewIntStat(16),
		Damage:   NewIntStat(5),
		Speed:    NewIntStat(7),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.8),
		CritRate: NewFloatStat(0.4),
	},
	{
		Name:     "Solid",
		Color:    "#3E5E88",
		Health:   NewIntStat(16),
		Damage:   NewIntStat(6),
		Speed:    NewIntStat(7),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.8),
		CritRate: NewFloatStat(0.3),
	},
	{
		Name:     "HTMX",
		Color:    "#3D72D7",
		Health:   NewIntStat(10),
		Damage:   NewIntStat(10),
		Speed:    NewIntStat(8),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.99),
		CritRate: NewFloatStat(0.4),
	},
	{
		Name:     "Datastar",
		Color:    "#BC4536",
		Health:   NewIntStat(8),
		Damage:   NewIntStat(11),
		Speed:    NewIntStat(9),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.99),
		CritRate: NewFloatStat(0.4),
	},
}

func chooseRandomFighter() Fighter {
	randomIndex := rand.IntN(len(fighterList))
	randomFighter := fighterList[randomIndex]
	// log.Printf("Randomly chose %v\n", randomFighter)
	return randomFighter
}

func chooseRandomFighterExclusive(excludedFighterName string) (Fighter, error) {
	swapIndex := -1
	for i := 0; i < len(fighterList); i += 1 {
		if fighterList[i].Name == excludedFighterName {
			swapIndex = i
			break
		}
	}
	if swapIndex == -1 {
		return Fighter{}, fmt.Errorf("error: Fighter name %s not found", excludedFighterName)
	}
	// Swap excluded fighter with first index
	temp := fighterList[0]
	fighterList[0] = fighterList[swapIndex]
	fighterList[swapIndex] = temp

	randomIndex := 1 + rand.IntN(len(fighterList)-1) // Choose from [1,1-len)
	randomFighter := fighterList[randomIndex]
	// log.Printf("Randomly chose %v, excluding %s\n", randomFighter, excludedFighterName)
	return randomFighter, nil
}

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
