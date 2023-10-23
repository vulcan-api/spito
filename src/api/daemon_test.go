package api

import "testing"

func TestGetDaemon(t *testing.T) {
	_, err := GetDaemon("dbus")

	if err != nil {
		t.Fatalf("Error occured when obtaining daemon data: %s", err)
	}
}
