package client

import "github.com/katya-spasova/music_player/player"
import (
	"errors"
	"net/http/httptest"
	"testing"
)

func TestGetAlive(t *testing.T) {
	ts := httptest.NewServer(player.InitService())
	defer ts.Close()
	defer player.ClearPlayer()

	cl := Client{host: ts.URL}
	found := cl.getAlive()
	expected := "I'm alive"
	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func checkStr(t *testing.T, expected string, found string) {
	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func TestFormUrl(t *testing.T) {
	cl := Client{host: "http://localhost:8765/"}
	checkStr(t, "http://localhost:8765/play/test_sounds%2Fbeep9.mp3", cl.formUrl("play", "test_sounds/beep9.mp3"))
	checkStr(t, "http://localhost:8765/save/abc.m3u", cl.formUrl("save", "abc.m3u"))
	checkStr(t, "http://localhost:8765/add/test_sounds%2Fbeep9.mp3", cl.formUrl("add", "test_sounds/beep9.mp3"))

	checkStr(t, "http://localhost:8765/pause", cl.formUrl("pause", "test_sounds/beep9.mp3"))
	checkStr(t, "http://localhost:8765/next", cl.formUrl("next", "djkfd"))
	checkStr(t, "http://localhost:8765/previous", cl.formUrl("previous", ""))
	checkStr(t, "http://localhost:8765/stop", cl.formUrl("stop", "djkfd"))
	checkStr(t, "http://localhost:8765/resume", cl.formUrl("resume", ""))
	checkStr(t, "http://localhost:8765/songinfo", cl.formUrl("songinfo", "djkfd"))
	checkStr(t, "http://localhost:8765/queueinfo", cl.formUrl("queueinfo", "djkfd"))
	checkStr(t, "http://localhost:8765/playlists", cl.formUrl("playlists", "djkfd"))
}

func TestDisplayMessage(t *testing.T) {
	c := ResponseContainer{Code: 0, Message: "Started Playing",
		Data: []string{"file1.mp3", "file two.flac", "/full/file3.ogg"}}
	expected := `Started Playing
file1.mp3
file two.flac
/full/file3.ogg`
	checkStr(t, expected, getDisplayMessage(c, nil))
}

func TestDisplayErrorMessage(t *testing.T) {
	c := ResponseContainer{Code: 1, Message: "Error Message",
		Data: []string{"file1.mp3", "file two.flac", "/full/file3.ogg"}}
	checkStr(t, "Error Message", getDisplayMessage(c, nil))
}

func TestDisplayError(t *testing.T) {
	c := ResponseContainer{Code: 1, Message: "Error Message",
		Data: []string{"file1.mp3", "file two.flac", "/full/file3.ogg"}}
	checkStr(t, "Error text", getDisplayMessage(c, errors.New("Error text")))
}
