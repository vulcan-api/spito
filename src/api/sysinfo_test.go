package api

import (
	"testing"
)

func TestGetCurrentDistro(t *testing.T) {
	distroName := GetCurrentDistro().Name

	if distroName == "" {
		t.Fatalf("ERROR! Couldn't detect your Linux distribution!")
	}
}

func TestGetCurrentInitSystem(t *testing.T) {
	initSystem, err := GetCurrentInitSystem()

	if err != nil || initSystem == "" {
		t.Fatalf("ERROR! Couldn't detect your init system!")
	}
}
