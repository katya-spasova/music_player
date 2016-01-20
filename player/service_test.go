package player

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func performCall(method string, url string) (string, error) {
	var res *http.Response
	var err error
	if method == "GET" {
		res, err = http.Get(url)
	} else if method == "POST" {
		res, err = http.Post(url, "text/plain", nil) //todo:
	} else if method == "PUT" {
		client := &http.Client{}
		request, err1 := http.NewRequest("PUT", url, nil)
		if err1 != nil {
			return "", err1
		}
		res, err = client.Do(request)
	}

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return "", err
	}
	found := string(body[:])
	return found, nil
}

func checkResult(method, url, expected string, t *testing.T) {
	found, err := performCall(method, url)
	if err != nil {
		t.Fatalf("Unexpected error found - %s", err.Error())
	}

	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func TestGetAlive(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	expected := "I'm alive"
	checkResult("GET", ts.URL, expected, t)
}

func TestPlay(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := `{"Code":0,"Message":"Success","Info":"Started playing",
	"Data":["test_sounds/beep9.mp3"]}`
	checkResult("PUT", url, expected, t)
}

func TestPlayDir(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	expected := `{"Code":0,"Message":"Success","Info":"Started playing",
	"Data":["test_sounds/beep9.mp3","test_sounds/beep28.mp3","test_sounds/beep36.mp3"]}`
	checkResult("PUT", url, expected, t)
}

func TestPlayPlaylist(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("sample_playlist")
	expected := `{"Code":0,"Message":"Success","Info":"Started playing",
	"Data":["test_sounds/beep9.mp3","test_sounds/beep28.mp3","test_sounds/beep36.mp3"]}`
	checkResult("PUT", url, expected, t)
}

func TestPlayNonExistingFile(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep1.mp3")
	expected := `{"Code":4,"Message":"File cannot be found"}`
	checkResult("PUT", url, expected, t)
}

func TestPlayInvalidFileFormat(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_broken/abc.txt")
	expected := `{"Code":7,"Message":"Format is not supported"}`
	checkResult("PUT", url, expected, t)
}

func TestPlayBrokenFile(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_broken/abc.txt")
	expected := `{"Code":2,"Message":"SoX failed to open input file"}`
	checkResult("PUT", url, expected, t)
}

func TestPause(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	url := ts.URL + "/pause"
	expected := `{"Code":0,"Message":"Success","Info":"Song is paused",
	"Data":["test_sounds/beep28.mp3"]}`

	checkResult("POST", url, expected, t)
}

func TestPauseNoPlayback(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/pause"
	expected := `{"Code":2,"Message":"Cannot pause. No song is playing"}`
	checkResult("POST", url, expected, t)
}

func TestResume(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	pause_url := ts.URL + "/pause"
	performCall("PUT", pause_url)

	url := ts.URL + "/resume"
	expected := `{"Code":0,"Message":"Success","Info":"Song is resumed",
	"Data":["test_sounds/beep28.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestResumeNoPlayback(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/resume"
	expected := `{"Code":9,"Message":"Cannot resume. No song was paused"}`
	checkResult("POST", url, expected, t)
}

func TestResumeNoPaused(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/resume"
	expected := `{"Code":9,"Message":"Cannot resume. No song was paused"}`
	checkResult("POST", url, expected, t)
}

func TestStop(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/stop"
	expected := `{"Code":0,"Message":"Success","Info":"Playback is stopped and cleaned","Data":[]}`
	checkResult("PUT", url, expected, t)
}

func TestStopNoPlayback(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/stop"
	expected := `{"Code":0,"Message":"Success","Info":"Playback is stopped and cleaned","Data":[]}`
	checkResult("PUT", url, expected, t)
}

func TestStopPaused(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	pause_url := ts.URL + "/pause"
	performCall("PUT", pause_url)

	url := ts.URL + "/stop"
	expected := `{"Code":0,"Message":"Success","Info":"Playback is stopped and cleaned","Data":[]}`
	checkResult("PUT", url, expected, t)
}

func TestNext(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := ts.URL + "/next"
	expected := `{"Code":0,"Message":"Success","Info":"Started playing",
	"Data":["test_sounds/beep28.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestNextNoNext(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/next"
	expected := `{"Code":10,"Message":"Cannot play next song. No next song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestNextNoPlayback(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/next"
	expected := `{"Code":10,"Message":"Cannot play next song. No next song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestPrevious(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)
	next_url := ts.URL + "/next"
	performCall("POST", next_url)

	url := ts.URL + "/previous"
	expected := `{"Code":0,"Message":"Success","Info":"Started playing",
	"Data":["test_sounds/beep9.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestPreviousNoPrevious(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/previous"
	expected := `{"Code":11,"Message":"Cannot play previous song. No previous song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestPreviousNoPlayback(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/previous"
	expected := `{"Code":11,"Message":"Cannot play previous song. No previous song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestGetCurrentSongInfo(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/songinfo"
	expected := `{"Code":0,"Message":"Success","Info":"The filename of the current song",
	"Data":["test_sounds/beep28.mp3"]}`
	checkResult("GET", url, expected, t)
}

func TestGetCurrentSongInfoNoPlayback(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/songinfo"
	expected := `{"Code":12,"Message":"There is no current song in the queue"}`
	checkResult("GET", url, expected, t)
}

func TestAdd(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := `{"Code":0,"Message":"Success","Info":"Added to queue",
	"Data":["test_sounds/beep9.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestAddDir(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_sounds")
	expected := `{"Code":0,"Message":"Success","Info":"Added to queue",
	"Data":["test_sounds/beep9.mp3","test_sounds/beep28.mp3","test_sounds/beep36.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestAddPlaylist(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("sample_playlist")
	expected := `{"Code":0,"Message":"Success","Info":"Added to queue",
	"Data":["test_sounds/beep9.mp3","test_sounds/beep28.mp3","test_sounds/beep36.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestAddNonExistingFile(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep1.mp3")
	expected := `{"Code":4,"Message":"File cannot be found"}`
	checkResult("POST", url, expected, t)
}

func TestAddInvalidFileFormat(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_broken/abc.txt")
	expected := `{"Code":7,"Message":"Format is not supported"}`
	checkResult("POST", url, expected, t)
}

func TestAddWithAvailableQueue(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	add_url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep9.mp3")
	performCall("POST", add_url)

	url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := `{"Code":0,"Message":"Success","Info":"Added to queue",
	"Data":["test_sounds/beep9.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestSaveAsPlaylist(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := ts.URL + "/save/" + url.QueryEscape("sample_playlist")
	expected := `{"Code":0,"Message":"Success","Info":"The queue is saved as a playlist",
	"Data":["sample_playlist"]}`
	checkResult("PUT", url, expected, t)
}

func TestSaveAsPlaylistNoPlayback(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/save/" + url.QueryEscape("sample_playlist")
	expected := `{"Code":14,"Message":"Queue is empty and cannot be saved as playlist"}`
	checkResult("PUT", url, expected, t)
}

func TestSaveAsPlaylistWrongName(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := ts.URL + "/save/" + url.QueryEscape("abc/sample_playlist")
	expected := `{"Code":13,"Message":"Cannot save playlist"}`
	checkResult("PUT", url, expected, t)
}

func TestListPlaylists(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/playlists"
	found, err := performCall("GET", url)
	if err != nil {
		t.Fatalf("Unexpected error found - %s", err.Error())
	}
	if !strings.Contains(found, `{"Code":0,"Message":"Success",
	"Info":"A list of all saved playlists","Data":["`) {
		t.Errorf("playlists returned wrong result")
	}
}

func TestGetQueueInfo(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := ts.URL + "/queueinfo"
	expected := `{"Code":0,"Message":"Success","Info":"Queue content",
	"Data":["test_sounds/beep9.mp3","test_sounds/beep28.mp3","test_sounds/beep36.mp3"]}`
	checkResult("GET", url, expected, t)
}

func TestGetQueueInfoEmpty(t *testing.T) {
	ts := httptest.NewServer(initService())
	defer ts.Close()
	defer clearPlayer()
	url := ts.URL + "/queueinfo"
	expected := `{"Code":15,"Message":"Cannot get queue info. Queue is empty"}`
	checkResult("GET", url, expected, t)
}
