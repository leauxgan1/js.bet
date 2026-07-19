package game

import (
	"fmt"
	"strings"
)

// / Audio struct where true players will play and false will be reset each update
type AudioPlayer struct {
	AttackPlaying  bool
	AbilityPlaying string
	BlockPlaying   bool
	DodgePlaying   bool
	CritPlaying    bool
	WinnerPlaying  bool
}

func (a *AudioPlayer) Stop() {
	a.AttackPlaying = false
	a.BlockPlaying = false
	a.DodgePlaying = false
	a.CritPlaying = false
	a.WinnerPlaying = false
}

// TODO make this a for loop
//
//	Convert audioplayer state into a recognizable string the client can use to play audio
func (a AudioPlayer) FormatAudioPlayer() string {
	builder := strings.Builder{}
	if a.AttackPlaying {
		builder.Write([]byte("attack,"))
	}
	if a.AbilityPlaying != "" {
		fmt.Fprintf(&builder, "ability-%s,", a.AbilityPlaying)
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
	if builder.Len() == 0 {
		return "none"
	}
	return builder.String()
}
