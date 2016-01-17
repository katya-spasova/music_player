package main

import (
	"testing"
)

func TestPlaySingleFile(t *testing.T) {
	player := Player{}
	player.init()
	player.playSingleFile("beep9.mp3")
	player.clear()
}
