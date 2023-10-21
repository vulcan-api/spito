package api

import "testing"

func TestGetDaemon(t *testing.T) {
	daemonInfo, err := GetDaemon("systemctl", "dbus")

	if daemonInfo.name == "" || daemonInfo.isActive == unknown || daemonInfo.isEnabled == unknown {
		t.Fatalf("Error occured when obtaining daemon data: %s", err)
	}
}
