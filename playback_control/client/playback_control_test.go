package client

import (
	"testing"
)

func TestGetAlive(t *testing.T) {
	found := GetAlive()
	expected := "I'm alive"
	if found != expected {
		t.Errorf("Expected\n---\n%s\n---\nbut found\n---\n%s\n---\n", expected, found)
	}
}
