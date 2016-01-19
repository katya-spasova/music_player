package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

const HOST = "http://localhost:8765/"

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

func clearPlayback() {
	url := HOST + "stop"
	performCall("PUT", url)
}

func TestGetAlive(t *testing.T) {
	expected := "I'm alive"
	checkResult("GET", HOST, expected, t)
}

func TestPlay(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep9.mp3\"]}"
	checkResult("PUT", url, expected, t)
	clearPlayback()
}

func TestPlayDir(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("test_sounds")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep9.mp3\",\"test_sounds/beep28.mp3\",\"test_sounds/beep36.mp3\"]}"
	checkResult("PUT", url, expected, t)
	clearPlayback()
}

func TestPlayPlaylist(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("sample_playlist")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep9.mp3\",\"test_sounds/beep28.mp3\",\"test_sounds/beep36.mp3\"]}"
	checkResult("PUT", url, expected, t)
	clearPlayback()
}

func TestPlayNonExistingFile(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("test_sounds/beep1.mp3")
	expected := "{\"Code\":4,\"Message\":\"File cannot be found\"}"
	checkResult("PUT", url, expected, t)
}

func TestPlayInvalidFileFormat(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("test_broken/abc.txt")
	expected := "{\"Code\":7,\"Message\":\"Format is not supported\"}"
	checkResult("PUT", url, expected, t)
}

func TestPlayBrokenFile(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("test_broken/abc.txt")
	expected := "{\"Code\":2,\"Message\":\"SoX failed to open input file\"}"
	checkResult("PUT", url, expected, t)
}

func TestPause(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	url := HOST + "pause"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Song is paused\",\"Data\":[\"test_sounds/beep28.mp3\"]}"

	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestPauseNoPlayback(t *testing.T) {
	url := HOST + "pause"
	expected := "{\"Code\":2,\"Message\":\"Cannot pause. No song is playing\"}"
	checkResult("POST", url, expected, t)
}

func TestResume(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	pause_url := HOST + "pause"
	performCall("PUT", pause_url)

	url := HOST + "resume"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Song is resumed\",\"Data\":[\"test_sounds/beep28.mp3\"]}"
	checkResult("POST", url, expected, t)
	clearPlayback()

}

func TestResumeNoPlayback(t *testing.T) {
	url := HOST + "resume"
	expected := "{\"Code\":9,\"Message\":\"Cannot resume. No song was paused\"}"
	checkResult("POST", url, expected, t)
}

func TestResumeNoPaused(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := HOST + "resume"
	expected := "{\"Code\":9,\"Message\":\"Cannot resume. No song was paused\"}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestStop(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := HOST + "stop"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Playback is stopped and cleaned\",\"Data\":[]}"
	checkResult("PUT", url, expected, t)
}

func TestStopNoPlayback(t *testing.T) {
	url := HOST + "stop"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Playback is stopped and cleaned\",\"Data\":[]}"
	checkResult("PUT", url, expected, t)
}

func TestStopPaused(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)
	pause_url := HOST + "pause"
	performCall("PUT", pause_url)

	url := HOST + "stop"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Playback is stopped and cleaned\",\"Data\":[]}"
	checkResult("PUT", url, expected, t)
}

func TestNext(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := HOST + "next"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep28.mp3\"]}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestNextNoNext(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := HOST + "next"
	expected := "{\"Code\":10,\"Message\":\"Cannot play next song. No next song in queue\"}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestNextNoPlayback(t *testing.T) {
	url := HOST + "next"
	expected := "{\"Code\":10,\"Message\":\"Cannot play next song. No next song in queue\"}"
	checkResult("POST", url, expected, t)
}

func TestPrevious(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)
	next_url := HOST + "next"
	performCall("POST", next_url)

	url := HOST + "previous"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep9.mp3\"]}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestPreviousNoPrevious(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := HOST + "previous"
	expected := "{\"Code\":11,\"Message\":\"Cannot play previous song. No previous song in queue\"}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestPreviousNoPlayback(t *testing.T) {
	url := HOST + "previous"
	expected := "{\"Code\":11,\"Message\":\"Cannot play previous song. No previous song in queue\"}"
	checkResult("POST", url, expected, t)
}

func TestGetCurrentSongInfo(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds/beep28.mp3")
	performCall("PUT", play_url)

	url := HOST + "songinfo"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"The filename of the current song\",\"Data\":[\"test_sounds/beep28.mp3\"]}"
	checkResult("GET", url, expected, t)
	clearPlayback()
}

func TestGetCurrentSongInfoNoPlayback(t *testing.T) {
	url := HOST + "songinfo"
	expected := "{\"Code\":12,\"Message\":\"There is no current song in the queue\"}"
	checkResult("GET", url, expected, t)
}

func TestAdd(t *testing.T) {
	url := HOST + "add/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Added to queue\",\"Data\":[\"test_sounds/beep9.mp3\"]}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestAddDir(t *testing.T) {
	url := HOST + "add/" + url.QueryEscape("test_sounds")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Added to queue\",\"Data\":[\"test_sounds/beep9.mp3\",\"test_sounds/beep28.mp3\",\"test_sounds/beep36.mp3\"]}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestAddPlaylist(t *testing.T) {
	url := HOST + "add/" + url.QueryEscape("sample_playlist")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Added to queue\",\"Data\":[\"test_sounds/beep9.mp3\",\"test_sounds/beep28.mp3\",\"test_sounds/beep36.mp3\"]}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestAddNonExistingFile(t *testing.T) {
	url := HOST + "add/" + url.QueryEscape("test_sounds/beep1.mp3")
	expected := "{\"Code\":4,\"Message\":\"File cannot be found\"}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestAddInvalidFileFormat(t *testing.T) {
	url := HOST + "add/" + url.QueryEscape("test_broken/abc.txt")
	expected := "{\"Code\":7,\"Message\":\"Format is not supported\"}"
	checkResult("POST", url, expected, t)
}

func TestAddWithAvailableQueue(t *testing.T) {
	add_url := HOST + "add/" + url.QueryEscape("test_sounds/beep9.mp3")
	performCall("POST", add_url)

	url := HOST + "add/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Added to queue\",\"Data\":[\"test_sounds/beep9.mp3\"]}"
	checkResult("POST", url, expected, t)
	clearPlayback()
}

func TestSaveAsPlaylist(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := HOST + "save/" + url.QueryEscape("sample_playlist")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"The queue is saved as a playlist\",\"Data\":[\"sample_playlist\"]}"
	checkResult("PUT", url, expected, t)
	clearPlayback()
}

func TestSaveAsPlaylistNoPlayback(t *testing.T) {
	url := HOST + "save/" + url.QueryEscape("sample_playlist")
	expected := "{\"Code\":14,\"Message\":\"Queue is empty and cannot be saved as playlist\"}"
	checkResult("PUT", url, expected, t)
}

func TestSaveAsPlaylistWrongName(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := HOST + "save/" + url.QueryEscape("abc/sample_playlist")
	expected := "{\"Code\":13,\"Message\":\"Cannot save playlist\"}"
	checkResult("PUT", url, expected, t)
	clearPlayback()
}

func TestListPlaylists(t *testing.T) {
	url := HOST + "playlists"
	found, err := performCall("GET", url)
	if err != nil {
		t.Fatalf("Unexpected error found - %s", err.Error())
	}
	if !strings.Contains(found, "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"A list of all saved playlists\",\"Data\":[") {
		t.Errorf("playlists returned wrong result")
	}
}

func TestGetQueueInfo(t *testing.T) {
	play_url := HOST + "play/" + url.QueryEscape("test_sounds")
	performCall("PUT", play_url)

	url := HOST + "queueinfo"
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Queue content\",\"Data\":[\"test_sounds/beep9.mp3\",\"test_sounds/beep28.mp3\",\"test_sounds/beep36.mp3\"]}"
	checkResult("GET", url, expected, t)
	clearPlayback()
}

func TestGetQueueInfoEmpty(t *testing.T) {
	url := HOST + "queueinfo"
	expected := "{\"Code\":15,\"Message\":\"Cannot get queue info. Queue is empty\"}"
	checkResult("GET", url, expected, t)
}
