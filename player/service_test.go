package main

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

const HOST = "http://localhost:8765/"

func TestGetAlive(t *testing.T) {
	response, err := http.Get(HOST)
	if err != nil {
		t.Fatalf("Unexpected error found - %s", err.Error())
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	found := string(body[:])

	expected := "I'm alive"
	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}

func TestPlay(t *testing.T) {
	filename := "test_sounds/beep9.mp3"
	url := HOST + "play/" + url.QueryEscape(filename)

	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, nil)
	res, err := client.Do(request)

	if err != nil {
		t.Fatalf("Unexpected error found - %s", err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf(err.Error())
	}
	found := string(body[:])
	expected := "Playing, test_sounds/beep9.mp3!"
	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}
