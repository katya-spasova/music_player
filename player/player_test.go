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
	checkIntFatal(t, Waiting, player.state.status)
	start := time.Now()
	player.playSingleFile("test_sounds/beep9.mp3", 0, nil)
	checkDuration(t, 0.9, 1.2, time.Since(start).Seconds())

	checkInt(t, Waiting, player.state.status)
	player.clear()
}

func TestSupportedTypes(t *testing.T) {
	fmt.Println("TestSupportedTypes")
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

func TestPlayFile(t *testing.T) {
	fmt.Println("TestPlayFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	start := time.Now()
	items, err := player.play("test_sounds/beep9.mp3")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkIntFatal(t, 1, len(items))
	checkStr(t, "test_sounds/beep9.mp3", items[0])
	checkInt(t, 1, len(player.state.queue))

	player.clear()
	checkDuration(t, 0.9, 1.2, time.Since(start).Seconds())
}

func checkIntFatal(t *testing.T, expected int, found int) {
	if found != expected {
		t.Fatalf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expected, found)
	}
}

func checkInt(t *testing.T, expected int, found int) {
	if found != expected {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expected, found)
	}
}

func checkStr(t *testing.T, expected string, found string) {
	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func checkDuration(t *testing.T, expectedMin float64, expectedMax float64, found float64) {
	if found < expectedMin || found > expectedMax {
		t.Errorf("Expected to play for around %f seconds, but played for %f seconds", expectedMin,
			found)
	}
}

func TestPlayerPlayDir(t *testing.T) {
	fmt.Println("TestPlayerPlayDir")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	start := time.Now()
	items, err := player.play("test_sounds")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkIntFatal(t, 3, len(items))
	checkStr(t, "test_sounds/beep28.mp3", items[0])
	checkStr(t, "test_sounds/beep36.mp3", items[1])
	checkStr(t, "test_sounds/beep9.mp3", items[2])
	checkInt(t, 3, len(player.state.queue))
	player.clear()
	checkDuration(t, 6.5, 6.8, time.Since(start).Seconds())
}

func TestPlayerPlaylist(t *testing.T) {
	fmt.Println("TestPlayerPlaylist")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	start := time.Now()
	items, err := player.play("sample_playlist.m3u")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkIntFatal(t, 3, len(items))
	checkStr(t, "test_sounds/beep9.mp3", items[0])
	checkStr(t, "test_sounds/beep28.mp3", items[1])
	checkStr(t, "test_sounds/beep36.mp3", items[2])
	checkInt(t, 3, len(player.state.queue))
	player.clear()
	checkDuration(t, 6.5, 6.8, time.Since(start).Seconds())
}

func TestPlayerPlayWrongFormat(t *testing.T) {
	fmt.Println("TestPlayerPlayWrongFormat")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	items, err := player.play("test_broken/abc.txt")
	if err == nil {
		t.Fatalf("Error expected")
	}
	checkStr(t, format_not_supported_msg, err.Error())
	checkInt(t, 0, len(items))
	checkInt(t, 0, len(player.state.queue))
	player.clear()
}

func TestPlayerPlayBrokenFile(t *testing.T) {
	fmt.Println("TestPlayerPlayBrokenFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	items, err := player.play("test_broken/no_music.mp3")
	if err == nil {
		t.Fatalf("Error expected")
	}
	checkStr(t, no_sox_in_msg, err.Error())
	checkInt(t, 1, len(items))
	checkInt(t, 1, len(player.state.queue))
	player.clear()
}
