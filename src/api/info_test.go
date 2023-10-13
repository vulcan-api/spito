package api

import (
	"testing"
)

func TestGetCurrentDistro(t *testing.T) {
	distroName := GetCurrentDistro().name

	if distroName == "" {
		t.Fatalf("ERROR! Couldn't detect your Linux distribution!")
	}
}
