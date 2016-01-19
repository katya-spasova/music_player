package player

import (
	"sync"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	player = Player{}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	player.clear()
}

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
		t.Errorf("Expected to play for at least\n---\n%d\n---\nbut played\n---\n%f\n---\n", 1,
			duration.Seconds())
	}
	player.clear()
}
