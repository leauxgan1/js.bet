package game

import (
	"fmt"
)

/// Audio struct where true players will play and false will be reset each update
type AudioPlayer struct {
	AttackPlaying bool
	BlockPlaying bool
	DodgePlaying bool
	CritPlaying bool
	WinnerPlaying bool
}

func (a *AudioPlayer) Stop() {
	a.AttackPlaying = false
	a.BlockPlaying = false
	a.DodgePlaying = false
	a.CritPlaying = false
	a.WinnerPlaying = false
}

// TODO make this a for loop
/// Convert audioplayer state into javascript able to be executed from the server, in which all on players play and all off players stop and reset
func (a AudioPlayer) FormatAudioPlayer() string {
	var command string = ""
	if a.AttackPlaying == true {
		command += FormatAudioPlayCommand("attack-player")
	} else {
		command += FormatAudioStopCommand("attack-player")
	}
	if a.BlockPlaying == true {
		command += FormatAudioPlayCommand("block-player")
	} else {
		command += FormatAudioStopCommand("block-player")
	}
	if a.DodgePlaying == true {
		command += FormatAudioPlayCommand("dodge-player")
	} else {
		command += FormatAudioStopCommand("dodge-player")
	}
	if a.CritPlaying == true {
		command += FormatAudioPlayCommand("crit-player")
	} else {
		command += FormatAudioStopCommand("crit-player")
	}
	if a.WinnerPlaying == true {
		command += FormatAudioPlayCommand("winner-player")
	} else {
		command += FormatAudioStopCommand("winner-player")
	}
	return command
}

func FormatAudioPlayCommand(id string) string {
	return fmt.Sprintf("document.querySelector('#%s').play();",id) 
}

func FormatAudioStopCommand(id string) string {
	return fmt.Sprintf("var p = document.querySelector('#%s');p.pause();p.currentTime = 0;",id) 
}

