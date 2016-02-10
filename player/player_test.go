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
	player = Player{wg: &wg}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	wg.Add(1)
	if player.state.status != Waiting {
		t.Errorf("Expected to status Waiting \n---\n%d\n---\nbut found \n---\n%d\n---\n", Waiting, player.state.status)
	}
	start := time.Now()
	player.playSingleFile("test_sounds/beep9.mp3", 0)
	if player.state.status != Playing {
		t.Errorf("Expected to status Playing \n---\n%d\n---\nbut found \n---\n%d\n---\n", Playing, player.state.status)
	}
	wg.Wait()
	duration := time.Since(start)
	expected := 0.9
	if duration.Seconds() < 0.9 {
		t.Errorf("Expected to play for at least\n---\n%d\n---\nbut played\n---\n%d\n---\n", expected,
			duration.Seconds())
	}
	if player.state.status != Waiting {
		t.Errorf("Expected to status Waiting \n---\n%d\n---\nbut found \n---\n%d\n---\n", Waiting, player.state.status)
	}
	player.clear()
}

func TestSupportedTypes(t *testing.T) {
	mp3Name := "abc.mp3"
	oggName := "abc.ogg"
	flacName := "abc.flac"
	wrongMp3 := "mp3"
	txtName := "abc.txt"
	emptyName := ""

	if !isSupportedType(mp3Name) {
		t.Errorf("Expected %s to be supported, but it's not", mp3Name)
	}

	if !isSupportedType(oggName) {
		t.Errorf("Expected %s to be supported, but it's not", oggName)
	}

	if !isSupportedType(flacName) {
		t.Errorf("Expected %s to be supported, but it's not", flacName)
	}

	if isSupportedType(wrongMp3) {
		t.Errorf("Expected %s NOT to be supported, but it is", wrongMp3)
	}

	if isSupportedType(txtName) {
		t.Errorf("Expected %s NOT to be supported, but it is", txtName)
	}

	if isSupportedType(emptyName) {
		t.Errorf("Expected %s NOT to be supported, but it is", emptyName)
	}
}
