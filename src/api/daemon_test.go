package api

import "testing"

func TestGetDaemon(t *testing.T) {
	daemon, err := GetDaemon("dbus")

	if err != nil {
		t.Fatalf("Error occured when obtaining daemon data: %s", err)
	}

	if daemon.Name == "" {
		t.Fatalf("Daemon name is empty")
	}
}
