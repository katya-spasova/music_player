package player

import (
	"fmt"
	"os"
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
	player.clear()
}

func TestPlaySingleFile(t *testing.T) {
	fmt.Println("TestPlaySingleFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	checkIntFatal(t, Waiting, player.state.status)
	start := time.Now()
	player.playSingleFile("test_sounds/beep9.mp3", 0, nil)
	checkDuration(t, 0.9, 1.2, time.Since(start).Seconds())

	checkInt(t, Waiting, player.state.status)
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
	defer player.clear()
	items, err := player.play("test_broken/abc.txt")
	if err == nil {
		t.Fatalf("Error expected")
	}
	checkStr(t, format_not_supported_msg, err.Error())
	checkInt(t, 0, len(items))
	checkInt(t, 0, len(player.state.queue))
}

func TestPlayerPlayBrokenFile(t *testing.T) {
	fmt.Println("TestPlayerPlayBrokenFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	items, err := player.play("test_broken/no_music.mp3")
	if err == nil {
		t.Fatalf("Error expected")
	}
	checkStr(t, no_sox_in_msg, err.Error())
	checkInt(t, 1, len(items))
	checkInt(t, 1, len(player.state.queue))
}

func TestAddPlayItemFile(t *testing.T) {
	fmt.Println("TestAddPlayItemFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()

	items, err := player.addPlayItem("test_sounds/beep9.mp3")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkInt(t, 1, len(items))
	checkInt(t, 1, len(player.state.queue))
	checkStr(t, "test_sounds/beep9.mp3", items[0])
}

func TestAddPlayItemDir(t *testing.T) {
	fmt.Println("TestAddPlayItemDir")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()

	items, err := player.addPlayItem("test_sounds")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkInt(t, 3, len(items))
	checkInt(t, 3, len(player.state.queue))
	checkStr(t, "test_sounds/beep28.mp3", items[0])
	checkStr(t, "test_sounds/beep36.mp3", items[1])
	checkStr(t, "test_sounds/beep9.mp3", items[2])
}

func TestAddPlayItemPlaylist(t *testing.T) {
	fmt.Println("TestAddPlayItemPlaylist")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()

	items, err := player.addPlayItem("sample_playlist.m3u")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkInt(t, 3, len(items))
	checkInt(t, 3, len(player.state.queue))
	checkStr(t, "test_sounds/beep9.mp3", items[0])
	checkStr(t, "test_sounds/beep28.mp3", items[1])
	checkStr(t, "test_sounds/beep36.mp3", items[2])
}

func TestAddPlayItemWrongFormat(t *testing.T) {
	fmt.Println("TestAddPlayItemWrongFormat")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()

	items, err := player.addPlayItem("test_broken/abc.txt")
	if err == nil {
		t.Fatalf("Error expected")
	}
	checkStr(t, format_not_supported_msg, err.Error())
	checkInt(t, 0, len(items))
	checkInt(t, 0, len(player.state.queue))
}

func TestAddPlayItemNotExisting(t *testing.T) {
	fmt.Println("TestAddPlayItemNotExisting")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()

	items, err := player.addPlayItem("abc.m3u")
	if err == nil {
		t.Fatalf("Error expected")
	}
	checkStr(t, file_not_found_msg, err.Error())
	checkInt(t, 0, len(items))
	checkInt(t, 0, len(player.state.queue))
}

func TestAddRegularFile(t *testing.T) {
	fmt.Println("TestAddRegularFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	items := player.addRegularFile("test_sounds/beep9.mp3")
	checkInt(t, 1, len(items))
	checkInt(t, 1, len(player.state.queue))
	checkStr(t, "test_sounds/beep9.mp3", items[0])
}

func TestAddRegularFilePlaylist(t *testing.T) {
	fmt.Println("TestAddRegularFilePlaylist")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	items := player.addRegularFile("playlists/sample_playlist.m3u")
	checkInt(t, 3, len(items))
	checkInt(t, 3, len(player.state.queue))
	checkStr(t, "test_sounds/beep9.mp3", items[0])
	checkStr(t, "test_sounds/beep28.mp3", items[1])
	checkStr(t, "test_sounds/beep36.mp3", items[2])
}

func TestAddRegularFileNotSupported(t *testing.T) {
	fmt.Println("TestAddRegularFileNotSupported")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	items := player.addRegularFile("test_brocken/abc.txt")
	checkInt(t, 0, len(items))
	checkInt(t, 0, len(player.state.queue))
}

func TestAddRegularFileNotExisting(t *testing.T) {
	fmt.Println("TestAddRegularFileNotSupported")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	items := player.addRegularFile("test_bro")
	checkInt(t, 0, len(items))
	checkInt(t, 0, len(player.state.queue))
}

func TestAddFile(t *testing.T) {
	fmt.Println("TestAddFile")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	item, err := player.addFile("test_sounds/beep9.mp3")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkStr(t, "test_sounds/beep9.mp3", item)
	checkInt(t, 1, len(player.state.queue))
}

func TestAddFileNotSupported(t *testing.T) {
	fmt.Println("TestAddFileNotSupported")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	items, err := player.addFile("test_broken/abc.txt")
	if err == nil {
		t.Errorf("Error expected")
	} else {
		checkStr(t, format_not_supported_msg, err.Error())
	}
	checkInt(t, 0, len(items))
	checkInt(t, 0, len(player.state.queue))
}

func TestPlayQueueNoTrim(t *testing.T) {
	fmt.Println("TestPlayQueueNoTrim")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	player.addFile("test_sounds/beep9.mp3")
	start := time.Now()
	player.playQueue(0, nil)
	checkDuration(t, 0.9, 1.1, time.Since(start).Seconds())
}

func TestPlayQueueTrim(t *testing.T) {
	fmt.Println("TestPlayQueueTrim")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	player.addFile("test_sounds/beep28.mp3")
	start := time.Now()
	player.playQueue(3.1, nil)
	checkDuration(t, 1.4, 1.7, time.Since(start).Seconds())
}

func TestSavePlaylistNoDir(t *testing.T) {
	fmt.Println("TestSavePlaylistNoDir")
	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	player.addRegularFile(player.playlistsDir + "sample_playlist.m3u")

	os.Rename(player.playlistsDir, "tmp")
	item, err := player.saveAsPlaylist("sample_playlist.m3u")
	if err != nil {
		t.Fatalf(err.Error())
	}
	checkStr(t, "sample_playlist.m3u", item)
	os.RemoveAll(player.playlistsDir)
	os.Rename("tmp", player.playlistsDir)
}

func TestListPlaylistNoDir(t *testing.T) {
	fmt.Println("TestListPlaylistNoDir")

	player = Player{playQueueMutex: &sync.Mutex{}}
	err := player.init()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer player.clear()
	player.addRegularFile(player.playlistsDir + "sample_playlist.m3u")
	//
	os.Rename(player.playlistsDir, "tmp/")

	_, err = player.listPlaylists()
	if err == nil {
		t.Fatalf("Error expected")
	}
	checkStr(t, playlist_not_found_msg, err.Error())
	os.Rename("tmp", player.playlistsDir)
}
