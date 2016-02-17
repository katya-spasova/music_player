package player

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func performCall(method string, url string) (string, error) {
	var res *http.Response
	var err error
	if method == "GET" {
		res, err = http.Get(url)
	} else if method == "POST" {
		res, err = http.Post(url, "text/plain", nil)
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
	time.Sleep(10 * time.Millisecond)
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
	fmt.Println("TestGetAlive")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	expected := "I'm alive"
	checkResult("GET", ts.URL, expected, t)
}

func TestPlay(t *testing.T) {
	fmt.Println("TestPlay")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := `{"Code":0,"Message":"Started playing","Data":["test_sounds/beep9.mp3"]}`
	checkResult("PUT", url, expected, t)
}

func TestPlayDir(t *testing.T) {
	fmt.Println("TestPlayDir")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	expected := `{"Code":0,"Message":"Started playing","Data":["test_sounds/beep28.mp3","test_sounds/beep36.mp3","test_sounds/beep9.mp3"]}`
	checkResult("PUT", url, expected, t)
}

func TestPlayPlaylist(t *testing.T) {
	fmt.Println("TestPlayPlaylist")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("sample_playlist.m3u")
	expected := `{"Code":0,"Message":"Started playing","Data":["test_sounds/beep9.mp3","test_sounds/beep28.mp3","test_sounds/beep36.mp3"]}`
	checkResult("PUT", url, expected, t)
}

func TestPlayNonExistingFile(t *testing.T) {
	fmt.Println("TestPlayNonExistingFile")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep1.mp3")
	expected := `{"Code":1,"Message":"File cannot be found"}`
	checkResult("PUT", url, expected, t)
}

func TestPlayInvalidFileFormat(t *testing.T) {
	fmt.Println("TestPlayNonExistingFile")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_broken/abc.txt")
	expected := `{"Code":1,"Message":"Format is not supported"}`
	checkResult("PUT", url, expected, t)
}

func TestPlayBrokenFile(t *testing.T) {
	fmt.Println("TestPlayBrokenFile")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/play/" + url.QueryEscape("test_broken/no_music.mp3")
	expected := `{"Code":1,"Message":"SoX failed to open input file"}`
	checkResult("PUT", url, expected, t)
}

func TestPause(t *testing.T) {
	fmt.Println("TestPause")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	url := ts.URL + "/pause"
	expected := `{"Code":0,"Message":"Song is paused","Data":["test_sounds/beep28.mp3"]}`

	checkResult("POST", url, expected, t)
}

func TestPauseNoPlayback(t *testing.T) {
	fmt.Println("TestPauseNoPlayback")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/pause"
	expected := `{"Code":1,"Message":"Cannot pause. No song is playing"}`
	checkResult("POST", url, expected, t)
}

func TestResume(t *testing.T) {
	fmt.Println("TestResume")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	time.Sleep(1 * time.Second)
	pause_url := ts.URL + "/pause"
	performCall("POST", pause_url)
	url := ts.URL + "/resume"
	expected := `{"Code":0,"Message":"Song is resumed","Data":["test_sounds/beep28.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestResumeNoPlayback(t *testing.T) {
	fmt.Println("TestResumeNoPlayback")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/resume"
	expected := `{"Code":1,"Message":"Cannot resume. No song was paused"}`
	checkResult("POST", url, expected, t)
}

func TestResumeNoPaused(t *testing.T) {
	fmt.Println("TestResumeNoPaused")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/resume"
	expected := `{"Code":1,"Message":"Cannot resume. No song was paused"}`
	checkResult("POST", url, expected, t)
}

func TestStop(t *testing.T) {
	fmt.Println("TestStop")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/stop"
	expected := `{"Code":0,"Message":"Playback is stopped and cleaned"}`
	checkResult("PUT", url, expected, t)
}

func TestStopNoPlayback(t *testing.T) {
	fmt.Println("TestStopNoPlayback")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/stop"
	expected := `{"Code":0,"Message":"Playback is stopped and cleaned"}`
	checkResult("PUT", url, expected, t)
}

func TestStopPaused(t *testing.T) {
	fmt.Println("TestStopPaused")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	pause_url := ts.URL + "/pause"
	performCall("PUT", pause_url)

	url := ts.URL + "/stop"
	expected := `{"Code":0,"Message":"Playback is stopped and cleaned"}`
	checkResult("PUT", url, expected, t)
}

func TestNext(t *testing.T) {
	fmt.Println("TestNext")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := ts.URL + "/next"
	expected := `{"Code":0,"Message":"Started playing","Data":["test_sounds/beep36.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestNextNoNext(t *testing.T) {
	fmt.Println("TestNextNoNext")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/next"
	expected := `{"Code":1,"Message":"Cannot play next song. No next song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestNextNoPlayback(t *testing.T) {
	fmt.Println("TestNextNoPlayback")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/next"
	expected := `{"Code":1,"Message":"Cannot play next song. No next song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestPrevious(t *testing.T) {
	fmt.Println("TestPrevious")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)
	next_url := ts.URL + "/next"
	performCall("POST", next_url)

	url := ts.URL + "/previous"
	expected := `{"Code":0,"Message":"Started playing","Data":["test_sounds/beep28.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestPreviousNoPrevious(t *testing.T) {
	fmt.Println("TestPreviousNoPrevious")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/previous"
	expected := `{"Code":1,"Message":"Cannot play previous song. No previous song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestPreviousNoPlayback(t *testing.T) {
	fmt.Println("TestPreviousNoPlayback")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/previous"
	expected := `{"Code":1,"Message":"Cannot play previous song. No previous song in queue"}`
	checkResult("POST", url, expected, t)
}

func TestGetCurrentSongInfo(t *testing.T) {
	fmt.Println("TestGetCurrentSongInfo")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := ts.URL + "/songinfo"
	expected := `{"Code":0,"Message":"The filename of the current song","Data":["test_sounds/beep28.mp3"]}`
	checkResult("GET", url, expected, t)
}

func TestGetCurrentSongInfoNoPlayback(t *testing.T) {
	fmt.Println("TestGetCurrentSongInfoNoPlayback")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/songinfo"
	expected := `{"Code":1,"Message":"There is no current song in the queue"}`
	checkResult("GET", url, expected, t)
}

func TestAdd(t *testing.T) {
	fmt.Println("TestAdd")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := `{"Code":0,"Message":"Added to queue","Data":["test_sounds/beep9.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestAddDir(t *testing.T) {
	fmt.Println("TestAddDir")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_sounds")
	expected := `{"Code":0,"Message":"Added to queue","Data":["test_sounds/beep28.mp3","test_sounds/beep36.mp3","test_sounds/beep9.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestAddPlaylist(t *testing.T) {
	fmt.Println("TestAddPlaylist")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("sample_playlist.m3u")
	expected := `{"Code":0,"Message":"Added to queue","Data":["test_sounds/beep9.mp3","test_sounds/beep28.mp3","test_sounds/beep36.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestAddNonExistingFile(t *testing.T) {
	fmt.Println("TestAddNonExistingFile")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep1.mp3")
	expected := `{"Code":1,"Message":"File cannot be found"}`
	checkResult("POST", url, expected, t)
}

func TestAddInvalidFileFormat(t *testing.T) {
	fmt.Println("TestAddInvalidFileFormat")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/add/" + url.QueryEscape("test_broken/abc.txt")
	expected := `{"Code":1,"Message":"Format is not supported"}`
	checkResult("POST", url, expected, t)
}

func TestAddWithAvailableQueue(t *testing.T) {
	fmt.Println("TestAddWithAvailableQueue")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	add_url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep9.mp3")
	performCall("POST", add_url)

	url := ts.URL + "/add/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := `{"Code":0,"Message":"Added to queue","Data":["test_sounds/beep9.mp3"]}`
	checkResult("POST", url, expected, t)
}

func TestSaveAsPlaylist(t *testing.T) {
	fmt.Println("TestSaveAsPlaylist")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("sample_playlist.m3u")
	performCall("PUT", play_url)

	url := ts.URL + "/save/" + url.QueryEscape("sample_playlist")
	expected := `{"Code":0,"Message":"The queue is saved as a playlist","Data":["sample_playlist.m3u"]}`
	checkResult("PUT", url, expected, t)
}

func TestSaveAsPlaylistNoPlayback(t *testing.T) {
	fmt.Println("TestSaveAsPlaylistNoPlayback")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/save/" + url.QueryEscape("sample_playlist")
	expected := `{"Code":1,"Message":"Queue is empty and cannot be saved as playlist"}`
	checkResult("PUT", url, expected, t)
}

func TestSaveAsPlaylistWrongName(t *testing.T) {
	fmt.Println("TestSaveAsPlaylistWrongName")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := ts.URL + "/save/" + url.QueryEscape("abc/sample_playlist")
	expected := `{"Code":1,"Message":"Cannot save playlist"}`
	checkResult("PUT", url, expected, t)
}

func TestListPlaylists(t *testing.T) {
	fmt.Println("TestListPlaylists")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/playlists"
	found, err := performCall("GET", url)
	if err != nil {
		t.Fatalf("Unexpected error found - %s", err.Error())
	}
	if !strings.Contains(found, `{"Code":0,"Message":"A list of all saved playlists","Data":["`) {
		t.Errorf("playlists returned wrong result")
	}
}

func TestGetQueueInfo(t *testing.T) {
	fmt.Println("TestGetQueueInfo")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	play_url := ts.URL + "/play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := ts.URL + "/queueinfo"
	expected := `{"Code":0,"Message":"Queue content","Data":["test_sounds/beep28.mp3","test_sounds/beep36.mp3","test_sounds/beep9.mp3"]}`
	checkResult("GET", url, expected, t)
}

func TestGetQueueInfoEmpty(t *testing.T) {
	fmt.Println("TestGetQueueInfoEmpty")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url := ts.URL + "/queueinfo"
	expected := `{"Code":1,"Message":"Cannot get queue info. Queue is empty"}`
	checkResult("GET", url, expected, t)
}

func TestWriteHttpResponse(t *testing.T) {
	fmt.Println("TestWriteHttpResponse")
	writer := httptest.NewRecorder()
	container := ResponseContainer{Code: 0, Message: "abc", Data: []string{"a", "b", "c"}}
	writeHttpResponse(writer, container)

	expectedCode := 200
	foundCode := writer.Code

	if foundCode != expectedCode {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expectedCode, foundCode)
	}

	expected := `{"Code":0,"Message":"abc","Data":["a","b","c"]}`
	found := writer.Body.String()

	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func TestFailWriteHttpResponse(t *testing.T) {
	fmt.Println("TestFailWriteHttpResponse")
	writer := httptest.NewRecorder()
	container := ResponseContainer{Code: 1, Message: "abc"}
	writeHttpResponse(writer, container)

	expectedCode := 200
	foundCode := writer.Code

	if foundCode != expectedCode {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expectedCode, foundCode)
	}

	expected := `{"Code":1,"Message":"abc"}`
	found := writer.Body.String()

	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func TestGetResponseContainerError(t *testing.T) {
	fmt.Println("TestGetResponseContainerError")
	data := []string{"a", "b", "c"}
	err := errors.New("Error message")

	container := getResponseContainer(data, err)

	if container.Data != nil {
		t.Errorf("Expected\n---\nnil\n---\nbut found\n---\n%v\n---\n", container.Data)
	}

	expected := "Error message"
	if container.Message != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, container.Message)
	}

	expectedCode := 1
	if container.Code != expectedCode {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expectedCode, container.Code)
	}
}

func TestGetResponseContainer(t *testing.T) {
	fmt.Println("TestGetResponseContainer")

	data := []string{"a", "b", "c"}
	var err error = nil

	container := getResponseContainer(data, err)

	if !reflect.DeepEqual(container.Data, data) {
		t.Errorf("Expected\n---\n%v\n---\nbut found\n---\n%v\n---\n", data, container.Data)
	}

	expected := ""
	if container.Message != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, container.Message)
	}

	expectedCode := 0
	if container.Code != expectedCode {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expectedCode, container.Code)
	}
}

func TestPlayerToServiceResponse(t *testing.T) {
	fmt.Println("TestPlayerToServiceResponse")
	writer := httptest.NewRecorder()

	playerToServiceResponse(writer, []string{"a", "b", "c"}, nil, "Success Message")

	expectedCode := 200
	foundCode := writer.Code

	if foundCode != expectedCode {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expectedCode, foundCode)
	}

	expected := `{"Code":0,"Message":"Success Message","Data":["a","b","c"]}`
	found := writer.Body.String()

	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func TestPlayerToServiceResponseError(t *testing.T) {
	fmt.Println("TestPlayerToServiceResponseError")
	writer := httptest.NewRecorder()

	playerToServiceResponse(writer, []string{"a", "b", "c"}, errors.New("Error"), "Success Message")

	expectedCode := 200
	foundCode := writer.Code

	if foundCode != expectedCode {
		t.Errorf("Expected\n---\n%d\n---\nbut found\n---\n%d\n---\n", expectedCode, foundCode)
	}

	expected := `{"Code":1,"Message":"Error"}`
	found := writer.Body.String()

	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func TestPlayAfterPlay(t *testing.T) {
	fmt.Println("TestPlayAfterPlay")
	ts := httptest.NewServer(InitService())
	defer ts.Close()
	defer ClearPlayer()
	url1 := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep36.mp3")
	expected := `{"Code":0,"Message":"Started playing","Data":["test_sounds/beep36.mp3"]}`
	checkResult("PUT", url1, expected, t)

	url2 := ts.URL + "/play/" + url.QueryEscape("test_sounds/beep28.mp3")
	expected2 := `{"Code":0,"Message":"Started playing","Data":["test_sounds/beep28.mp3"]}`
	checkResult("PUT", url2, expected2, t)
}
