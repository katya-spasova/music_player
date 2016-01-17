package main

import (
	"io/ioutil"
	"net/http"
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
