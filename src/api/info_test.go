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
