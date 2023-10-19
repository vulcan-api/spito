package api

import (
	"testing"
)

func TestGetDaemon(t *testing.T) {
	daemonInfo, err := GetDaemon("dbus", "systemctl")

	if err != nil {
		t.Fatalf("Errors occured during obtaining daemon info: %s", err)
	}

	t.Logf("Name: %s, IsActive: %s, IsEnabled: %s", daemonInfo.name, daemonInfo.isActive, daemonInfo.isEnabled)
}
