package api

import "testing"

func TestGetDaemon(t *testing.T) {
	daemon, err := GetDaemon("ssh")

	t.Log(daemon.IsActive)
	t.Log(daemon.IsEnabled)
	t.Log(daemon.InitLevel)

	if err != nil {
		t.Fatalf("Error occured when obtaining daemon data: %s", err)
	}

	if daemon.Name == "" {
		t.Fatalf("Daemon name is empty")
	}
}
