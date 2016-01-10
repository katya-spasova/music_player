package main

import (
	"testing"
)

func TestGetAlive(t *testing.T) {
	found := getAlive()
	expected := "I'm alive"
	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}
