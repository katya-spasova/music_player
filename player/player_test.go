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
	if player.state.status != Waiting {
		t.Errorf("Expected to status Waiting \n---\n%d\n---\nbut found \n---\n%f\n---\n", Waiting, player.state.status)
	}
	start := time.Now()
	player.playSingleFile("test_sounds/beep9.mp3", 0)
	if player.state.status != Playing {
		t.Errorf("Expected to status Playing \n---\n%d\n---\nbut found \n---\n%f\n---\n", Playing, player.state.status)
	}
	wg.Wait()
	duration := time.Since(start)
	expected := 0.9
	if duration.Seconds() < 0.9 {
		t.Errorf("Expected to play for at least\n---\n%d\n---\nbut played\n---\n%f\n---\n", expected,
			duration.Seconds())
	}
	if player.state.status != Waiting {
		t.Errorf("Expected to status Waiting \n---\n%d\n---\nbut found \n---\n%f\n---\n", Waiting, player.state.status)
	}
	player.clear()
}
