package game

import (
	"strings"
)

// / Audio struct where true players will play and false will be reset each update
type AudioPlayer struct {
	AttackPlaying        bool
	BlockPlaying         bool
	DodgePlaying         bool
	CritPlaying          bool
	WinnerPlaying        bool
	AbilityDamagePlaying bool
	AbilityBuffPlaying   bool
	AbilityDebuffPlaying bool
	AbilityTrickPlaying  bool
}

func (a *AudioPlayer) Stop() {
	a.AttackPlaying = false
	a.BlockPlaying = false
	a.DodgePlaying = false
	a.CritPlaying = false
	a.WinnerPlaying = false
	a.AbilityDamagePlaying = false
	a.AbilityBuffPlaying = false
	a.AbilityDebuffPlaying = false
	a.AbilityTrickPlaying = false
}

// TODO make this a for loop
//
//	Convert audioplayer state into a recognizable string the client can use to play audio
func (a AudioPlayer) FormatAudioPlayer() string {
	builder := strings.Builder{}
	// Combat sounds
	if a.AttackPlaying {
		builder.Write([]byte("attack,"))
	}
	if a.BlockPlaying {
		builder.Write([]byte("block,"))
	}
	if a.DodgePlaying {
		builder.Write([]byte("dodge,"))
	}
	if a.CritPlaying {
		builder.Write([]byte("crit,"))
	}
	if a.WinnerPlaying {
		builder.Write([]byte("winner,"))
	}
	// Ability Sounds
	if a.AbilityDamagePlaying {
		builder.Write([]byte("ability-damage,"))
	}
	if a.AbilityBuffPlaying {
		builder.Write([]byte("ability-buff,"))
	}
	if a.AbilityDebuffPlaying {
		builder.Write([]byte("ability-debuff,"))
	}
	if a.AbilityTrickPlaying {
		builder.Write([]byte("ability-trick,"))
	}
	if builder.Len() == 0 {
		return "none"
	}
	return builder.String()
}
