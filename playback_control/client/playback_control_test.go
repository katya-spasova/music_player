package client

import "github.com/katya-spasova/music_player/player"
import (
	"errors"
	"net/http/httptest"
	"testing"
)

var playlistsDir = "test_playlists"

func TestGetAlive(t *testing.T) {
	ts := httptest.NewServer(player.InitService(playlistsDir))
	defer ts.Close()
	defer player.WaitEnd()

	cl := Client{Host: ts.URL + "/"}
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

func checkInt(t *testing.T, expected int, found int) {
	if found != expected {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expected, found)
	}
}

func TestFormUrl(t *testing.T) {
	cl := Client{Host: "http://localhost:8765/"}
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

func TestPerformCallPut(t *testing.T) {
	ts := httptest.NewServer(player.InitService(playlistsDir))
	defer ts.Close()
	defer player.WaitEnd()

	responseContainer, err := performCall("PUT", ts.URL+"/play/"+escape("../../player/test_sounds/beep9.mp3"))
	if err != nil {
		t.Errorf(err.Error())
	}

	checkInt(t, 0, responseContainer.Code)
}

func TestPerformCallGet(t *testing.T) {
	ts := httptest.NewServer(player.InitService(playlistsDir))
	defer ts.Close()
	defer player.WaitEnd()

	responseContainer, err := performCall("GET", ts.URL+"/songinfo")
	if err != nil {
		t.Errorf(err.Error())
	}

	checkInt(t, 1, responseContainer.Code)
}

func TestPerformCallPost(t *testing.T) {
	ts := httptest.NewServer(player.InitService(playlistsDir))
	defer ts.Close()
	defer player.WaitEnd()

	responseContainer, err := performCall("POST", ts.URL+"/pause")
	if err != nil {
		t.Errorf(err.Error())
	}

	checkInt(t, 1, responseContainer.Code)
}

func TestPerformCallError(t *testing.T) {
	ts := httptest.NewServer(player.InitService(playlistsDir))
	defer ts.Close()
	defer player.WaitEnd()

	_, err := performCall("GET", ts.URL+"/pause")
	if err == nil {
		t.Error("Error Expected")
	}
}

func TestPerformAction(t *testing.T) {
	ts := httptest.NewServer(player.InitService(playlistsDir))
	defer ts.Close()
	defer player.WaitEnd()

	cl := Client{Host: ts.URL + "/"}
	found := cl.PerformAction("play", "../../player/test_sounds/beep9.mp3")
	expected := `Started playing
beep9.mp3`

	checkStr(t, expected, found)
}

func TestEscape(t *testing.T) {
	found := escape(`/abc cde\fgh.ijk`)
	checkStr(t, "%2Fabc%20cde%5Cfgh.ijk", found)
}

func TestIsLocalhost(t *testing.T) {
	cl := Client{Host: "http://localhost:8765"}
	if !cl.isLocalhostCall() {
		t.Error("http://localhost:8765 is expected to be localhost")
	}

	cl = Client{Host: "https://localhost:8765"}
	if !cl.isLocalhostCall() {
		t.Error("https://localhost:8765 is expected to be localhost")
	}

	cl = Client{Host: "http://127.0.0.1:8765"}
	if !cl.isLocalhostCall() {
		t.Error("http://127.0.0.1:8765 is expected to be localhost")
	}

	cl = Client{Host: "https://127.0.0.1:8765"}
	if !cl.isLocalhostCall() {
		t.Error("https://127.0.0.1:8765 is expected to be localhost")
	}

	cl = Client{Host: "http://192.168.0.1:8765"}
	if cl.isLocalhostCall() {
		t.Error("http://192.168.0.1:8765 is NOT expected to be localhost")
	}

	cl = Client{Host: "http://google.com"}
	if cl.isLocalhostCall() {
		t.Error("http://google.com is NOT expected to be localhost")
	}
}
