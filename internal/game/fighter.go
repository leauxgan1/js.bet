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
	ABILITYUSING
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
	Name        string
	Description string
	Timer       IntStat
	InvokeFunc  func(self *Fighter, other *Fighter)
}

type Fighter struct {
	Name        string       // Name of framework/library
	Color       string       // Color of logo
	Health      IntStat      // Represents how much of an "industry standard" the framework/library is / likelihood to stick around in the future
	Damage      IntStat      // Represents how consistently useful the framework/library is for common tasks
	Speed       IntStat      // Represents the overall performance under load and scalability of the framework/library, causes fighter to act sooner
	Accuracy    FloatStat    // Represents how simple the library/frame work is / how easy it is to get it right at first (opposite of footguns), causes less misses
	CritRate    FloatStat    // Represents how suprisingly useful or versatile the framework/library is in niche situations
	AttackTimer IntStat      // Time before next action of fighter, reduced by speed each turn
	State       FighterState // Current state of fighter, used for animations
	Abilities   []Ability    // Abilities which may apply status effects to fighters
	Effects     []Effect     // Effects which are applied by abilities and tick down over time
}

func (f *Fighter) Reset() *Fighter {
	f.State = READY
	f.Health.Value = f.Health.MaxValue
	for i := 0; i < len(f.Abilities); i++ {
		f.Abilities[i].Timer.Value = f.Abilities[i].Timer.MaxValue
	}
	f.Effects = make([]Effect, 0, 3)
	return f
}

/* Ability ideas:

React -> Virtual DOM: Increase Speed but reduce damage output slightly, I am inevitable...: Deal damage based on popularity [X]

Vue -> Second-most-loved: Heal a small amount, Vapor mode: Increases Speed and Damage slightly []

Solid -> Signals, signals everywhere...: Gain accuracy and speed []

Svelte -> Compile: Increase speed, Most-loved: Heal a moderate amount []

HTMX -> Out of touch: Increase dodge / Reduce Accuracy, Resilience: Deal damage based on defense []

Datastar -> Greedy: Lose some health / Gain Damage buff []

JQuery -> Old, not forgotten: Deal damage equal to max health []

*/

var fighterList = [...]Fighter{
	{
		Name:        "JQuery",
		Color:       "#0769AD",
		Health:      NewIntStat(40),
		Damage:      NewIntStat(4),
		Speed:       NewIntStat(8),
		AttackTimer: NewIntStat(20),
		Accuracy:    NewFloatStat(0.5),
		CritRate:    NewFloatStat(0.0),
		Abilities: []Ability{
			{
				Name:        "Old But Not Forgotten",
				Description: "",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
				Timer: NewIntStat(2),
			},
		},
		Effects: make([]Effect, 0, 3),
	},
	{
		Name:        "React",
		Color:       "#58C4DC",
		Health:      NewIntStat(30),
		Damage:      NewIntStat(5),
		Speed:       NewIntStat(4),
		AttackTimer: NewIntStat(20),
		Accuracy:    NewFloatStat(0.6),
		CritRate:    NewFloatStat(0.2),
		Abilities: []Ability{
			{
				Name:        "Virtual DOM",
				Description: "Slows everything down",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					// self.Damage.Value -= 2
					// self.Speed.Value += 2
					lastSpeed := other.Speed.Value
					slow := Slow{NewIntStat(10), 2, lastSpeed}
					other.Effects = append(other.Effects, &slow)
					slow.OnApply(other) // Don't forget to run onApply!

				},
				Timer: NewIntStat(1),
			},
			{
				Name:        "I am inevitable...",
				Description: "Crushes competition mainly due to inertia",
				InvokeFunc: func(self *Fighter, other *Fighter) {
					self.Damage.MaxValue *= 2
					self.Damage.Value = self.Damage.MaxValue
				},
				Timer: NewIntStat(1),
			},
		},
		Effects: make([]Effect, 0, 3),
	},
	{
		Name:        "Vue",
		Color:       "#00C180",
		Health:      NewIntStat(25),
		Damage:      NewIntStat(5),
		Speed:       NewIntStat(6),
		AttackTimer: NewIntStat(20),
		Accuracy:    NewFloatStat(0.8),
		CritRate:    NewFloatStat(0.3),
		Abilities: []Ability{
			{
				Name:        "Second most loved, btw!",
				Description: "",
				Timer:       NewIntStat(1),
				InvokeFunc: func(self *Fighter, other *Fighter) {
					self.Health.Value = min(self.Health.Value+10, self.Health.MaxValue)
				},
			},
		},
		Effects: make([]Effect, 0, 3),
	},
	{
		Name:        "Svelte",
		Color:       "#FF5018",
		Health:      NewIntStat(26),
		Damage:      NewIntStat(5),
		Speed:       NewIntStat(7),
		AttackTimer: NewIntStat(20),
		Accuracy:    NewFloatStat(0.8),
		CritRate:    NewFloatStat(0.4),
		Abilities: []Ability{
			{
				Name:        "Most Loved Framework, btw",
				Description: "",
				Timer:       NewIntStat(1),
				InvokeFunc: func(self *Fighter, other *Fighter) {
					self.Health.Value = min(self.Health.Value+10, self.Health.MaxValue)
				},
			},
		},
		Effects: make([]Effect, 0, 3),
	},
	{
		Name:        "Solid",
		Color:       "#3E5E88",
		Health:      NewIntStat(26),
		Damage:      NewIntStat(6),
		Speed:       NewIntStat(7),
		AttackTimer: NewIntStat(20),
		Accuracy:    NewFloatStat(0.8),
		CritRate:    NewFloatStat(0.3),
		Abilities: []Ability{
			{
				Name:        "Go my signals...",
				Description: "",
				Timer:       NewIntStat(2),
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
			},
		},
		Effects: make([]Effect, 0, 3),
	},
	{
		Name:        "HTMX",
		Color:       "#3D72D7",
		Health:      NewIntStat(20),
		Damage:      NewIntStat(10),
		Speed:       NewIntStat(8),
		AttackTimer: NewIntStat(20),
		Accuracy:    NewFloatStat(0.99),
		CritRate:    NewFloatStat(0.4),
		Abilities: []Ability{
			{
				Name:        "Web 1.0 Larp",
				Description: "",
				Timer:       NewIntStat(25),
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
			},
			{
				Name:        "Out of touch",
				Description: "",
				Timer:       NewIntStat(2),
				InvokeFunc: func(self *Fighter, other *Fighter) {
					other.Health.Value -= self.Health.MaxValue
				},
			},
		},
		Effects: make([]Effect, 0, 3),
	},
	{
		Name:        "Datastar",
		Color:       "#BC4536",
		Health:      NewIntStat(18),
		Damage:      NewIntStat(11),
		Speed:       NewIntStat(9),
		AttackTimer: NewIntStat(20),
		Accuracy:    NewFloatStat(0.99),
		CritRate:    NewFloatStat(0.4),
		Abilities: []Ability{
			{
				Name:        "Greedy Dev",
				Description: "",
				Timer:       NewIntStat(1),
				InvokeFunc: func(self *Fighter, other *Fighter) {
					//
				},
			},
		},
		Effects: make([]Effect, 0, 3),
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
	StepDuration()
	GetDuration() int
	OnApply(f *Fighter)
	OnTick(f *Fighter)
	OnRemove(f *Fighter)
}

type Slow struct {
	duration   IntStat
	SlowAmount int
	LastSpeed  int
}

func (s *Slow) StepDuration() {
	s.duration.Value -= 1
}

func (s *Slow) GetDuration() int {
	return s.duration.Value
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
