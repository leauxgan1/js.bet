package game

import (
	"fmt"
)

type AudioID int 
const (
	_ = iota
	AUDIO_ATTACK
	AUDIO_BLOCK
	AUDIO_DODGE
	AUDIO_CRIT
)

func FormatAudioPlayer(id AudioID) string {
	switch id {
		case AUDIO_ATTACK:
			return "attack_player"
		case AUDIO_BLOCK:
			return "block_player"
		case AUDIO_DODGE:
			return "dodge_player"
		case AUDIO_CRIT:
			return "crit_player"
	}
	return ""
}

func FormatAudioCommand(id AudioID) string {
	return fmt.Sprintf("document.querySelector('#%s').play();", FormatAudioPlayer(id))
}

func FormatStopAllAudio() string {
	var finalCommand = ""
	for i := range AUDIO_CRIT {
		finalCommand += fmt.Sprintf("let sound = document.querySelector('#%s); sound.pause(); sound.currentTime = 0;",FormatAudioPlayer(AudioID(i)))
	}
	return finalCommand
}
