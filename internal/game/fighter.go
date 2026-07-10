package game

import (
	"fmt"
	"math/rand/v2"
)

type FighterState uint

const (
	_ = iota
	READY
	DEFENDING
	DODGING
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

type Ability struct {
	Name       string
	InvokeFunc func(self *Fighter, other *Fighter)
	Timer      IntStat
}

type Fighter struct {
	Name      string       // Name of framework/library
	Color     string       // Color of logo
	Health    IntStat      // Represents how much of an "industry standard" the framework/library is / likelihood to stick around in the future
	Damage    IntStat      // Represents how consistently useful the framework/library is for common tasks
	Speed     IntStat      // Represents the overall performance under load and scalability of the framework/library, causes fighter to act sooner
	Accuracy  FloatStat    // Represents how simple the library/frame work is / how easy it is to get it right at first (opposite of footguns), causes less misses
	CritRate  FloatStat    // Represents how suprisingly useful or versatile the framework/library is in niche situations
	Timer     IntStat      // Time before next action of fighter, reduced by speed each turn
	State     FighterState // Current state of fighter, used for animations
	Abilities [4]Ability   // Abilities which may apply status effects to fighters
	Effects   []Effect     // Effects which are applied by abilities and tick down over time
}

func (f *Fighter) Reset() *Fighter {
	f.State = READY
	f.Health.Value = f.Health.MaxValue
	return f
}

/* Ability ideas:

React -> Virtual DOM: Increase Speed but reduce damage output slightly, I am inevitable...:
Vue -> Second-most-loved: Heal a small amount, Vapor mode: Increases Speed and Damage slightly
Solid -> Signals, signals everywhere...: Gain accuracy and speed
Svelte -> Compile: Increase speed, Most-loved: Heal a moderate amount
HTMX -> Out of touch: Increase dodge / Reduce Accuracy, Resilience: Deal damage based on defense
Datastar -> Greedy: Lose some health / Gain Damage buff
JQuery -> Old, not forgotten: Deal damage equal to max health

*/

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
		Abilities: [4]Ability{
			{
				Name: "Old But Not Forgotten",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
				Timer: NewIntStat(25),
			},
		},
		Effects: make([]Effect, 0),
	},
	{
		Name:     "React",
		Color:    "#58C4DC",
		Health:   NewIntStat(20),
		Damage:   NewIntStat(5),
		Speed:    NewIntStat(4),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.6),
		CritRate: NewFloatStat(0.2),
		Abilities: [4]Ability{
			{
				Name: "I am inevitable...",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					self.Damage.MaxValue *= 2
					self.Damage.Value = self.Damage.MaxValue
				},
				Timer: NewIntStat(15),
			},
		},
		Effects: make([]Effect, 1),
	},
	{
		Name:     "Vue",
		Color:    "#00C180",
		Health:   NewIntStat(15),
		Damage:   NewIntStat(5),
		Speed:    NewIntStat(6),
		Timer:    NewIntStat(20),
		Accuracy: NewFloatStat(0.8),
		CritRate: NewFloatStat(0.3),
		Abilities: [4]Ability{
			{
				Name: "Second most loved, btw!",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					self.Health.Value = min(self.Health.Value+10, self.Health.MaxValue)
				},
				Timer: NewIntStat(10),
			},
		},
		Effects: make([]Effect, 1),
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
		Abilities: [4]Ability{
			{
				Name: "Most Loved Framework, btw",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					self.Health.Value = min(self.Health.Value+10, self.Health.MaxValue)
				},
				Timer: NewIntStat(10),
			},
		},
		Effects: make([]Effect, 1),
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
		Abilities: [4]Ability{
			{
				Name: "Go my signals...",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
				Timer: NewIntStat(25),
			},
		},
		Effects: make([]Effect, 1),
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
		Abilities: [4]Ability{
			{
				Name: "Web 1.0 Larp",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
				Timer: NewIntStat(25),
			},
			{
				Name: "Out of touch",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
				Timer: NewIntStat(25),
			},
		},
		Effects: make([]Effect, 1),
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
		Abilities: [4]Ability{
			{
				Name: "Greedy Dev",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					//
				},
				Timer: NewIntStat(10),
			},
		},
		Effects: make([]Effect, 1),
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
	for i, fighter := range fighterList {
		if fighter.Name == excludedFighterName {
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

// Effects to apply from abilities to self or an opponent fighter

type Effect interface {
	OnApply(f *Fighter)
	OnTick(f *Fighter)
	OnRemove(f *Fighter)
}

type Slow struct {
	SlowAmount int
	LastSpeed  int
}

func (s *Slow) OnApply(f *Fighter) {
	s.LastSpeed = f.Speed.Value
	f.Speed.Value -= s.SlowAmount
	if f.Speed.Value < 0 {
		f.Speed.Value = 0
	}
}
func (s *Slow) OnTick(f *Fighter) {
	// Do nothing
}
func (s *Slow) OnRemove(f *Fighter) {
	f.Speed.Value = s.LastSpeed
}
