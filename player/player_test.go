package player

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	fmt.Println("TestInit")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	player.release()
	player.clear()
}

func TestPlaySingleFile(t *testing.T) {
	fmt.Println("TestPlaySingleFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	if player.state.status != Waiting {
		t.Errorf("Expected status Waiting \n---\n%d\n---\nbut found \n---\n%d\n---\n", Waiting, player.state.status)
	}
	start := time.Now()
	player.playSingleFile("test_sounds/beep9.mp3", 0, nil)
	duration := time.Since(start)
	expected := 0.9
	if duration.Seconds() < 0.9 {
		t.Errorf("Expected to play for at least\n---\n%d\n---\nbut played\n---\n%d\n---\n", expected,
			duration.Seconds())
	}
	fmt.Println(player.state.status)
	if player.state.status != Waiting {
		t.Errorf("Expected status Waiting \n---\n%d\n---\nbut found \n---\n%d\n---\n", Waiting, player.state.status)
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
