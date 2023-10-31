package api

import (
	"testing"
)

func TestGetDistro(t *testing.T) {
	distroName := GetDistro().Name

	if distroName == "" {
		t.Fatalf("ERROR! Couldn't detect your Linux distribution!")
	}
}

func TestGetInitSystem(t *testing.T) {
	initSystem, err := GetInitSystem()

	if err != nil || initSystem == UNKNOWN {
		t.Fatalf("ERROR! Couldn't detect your init system!")
	}
}
