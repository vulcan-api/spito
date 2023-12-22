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

	if daemon.RunLevel == "" {
		t.Fatalf("Run level is empty")
	}

	if !daemon.IsActive {
		t.Fatalf("Daemon is inactive")
	}

	if !daemon.IsEnabled {
		t.Fatalf("Daemon is not enabled")
	}

	t.Logf("\nData obtained about '%s' daemon:", daemon.Name)
	t.Logf("\nIs active: %t", daemon.IsActive)
	t.Logf("\nIs enabled: %t", daemon.IsEnabled)
	t.Logf("\nRun level: %s", daemon.RunLevel)
}
