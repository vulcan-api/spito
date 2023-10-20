package api

import (
	"testing"
)

func TestGetDaemon(t *testing.T) {
	daemonInfo, err := GetDaemon("dbus", "systemctl")

	if daemonInfo.name == "" || daemonInfo.isActive == unknown || daemonInfo.isEnabled == unknown {
		t.Fatalf("There's a problem when executing your init system: %s", err)
	}

	if daemonInfo.isEnabled == notFound {
		t.Fatalf("Daemon is not installed: %s", err)
	}

	t.Logf("Name: %s, IsActive: %d, IsEnabled: %d", daemonInfo.name, daemonInfo.isActive, daemonInfo.isEnabled)
}
