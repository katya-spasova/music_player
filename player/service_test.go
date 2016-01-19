package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
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

func TestGetAlive(t *testing.T) {
	expected := "I'm alive"
	checkResult("GET", HOST, expected, t)
}

func TestPlay(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("test_sounds/beep9.mp3")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep9.mp3\"]}"
	checkResult("PUT", url, expected, t)
}

func TestPlayDir(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("test_sounds")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep9.mp3\",\"test_sounds/beep28.mp3\",\"test_sounds/beep36.mp3\"]}"
	checkResult("PUT", url, expected, t)
}

func TestPlayPlaylist(t *testing.T) {
	url := HOST + "play/" + url.QueryEscape("sample_playlist")
	expected := "{\"Code\":0,\"Message\":\"Success\",\"Info\":\"Started playing\",\"Data\":[\"test_sounds/beep9.mp3\",\"test_sounds/beep28.mp3\",\"test_sounds/beep36.mp3\"]}"
	checkResult("PUT", url, expected, t)
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
