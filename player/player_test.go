package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPlaySingleFile(t *testing.T) {
	wg := sync.WaitGroup{}
	player = Player{WG: &wg}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	wg.Add(1)
	start := time.Now()
	player.playSingleFile("test_sounds/beep9.mp3")
	wg.Wait()
	duration := time.Since(start)
	if duration.Seconds() < 1 {
		fmt.Println("played file for ", duration.Seconds())
		t.Errorf("Could not play the sample file")
	}
	player.clear()
}
