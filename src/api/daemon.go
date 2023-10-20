package api

import (
	"errors"
	"os/exec"
	"strings"
)

const systemd = "systemctl"

type DaemonActive int
type DaemonEnabled int

const unknown = -1

const (
	active   DaemonActive = iota
	inactive              // what's funny systemd treats unrecognized daemons as inactive
	activating
	deactivating
	failed
)

const (
	enabled DaemonEnabled = iota
	disabled
	static
	notFound
)

type Daemon struct {
	name      string
	isActive  DaemonActive
	isEnabled DaemonEnabled
}

func getActivityIota(activity string) DaemonActive {
	switch activity {
	case "active":
		return active
	case "inactive":
		return inactive
	case "activating":
		return activating
	case "deactivating":
		return deactivating
	case "failed":
		return failed
	default:
		return unknown
	}
}

func getEnabledIota(isEnabled string) DaemonEnabled {
	switch isEnabled {
	case "isEnabled":
		return enabled
	case "disabled":
		return disabled
	case "static":
		return static
	case "not-found":
		return notFound
	default:
		return unknown
	}
}

func execSystemctl(command string, daemon string) (string, error) {
	cmd := exec.Command(systemd, command, daemon)
	stdout, err := cmd.Output()
	return strings.TrimSpace(string(stdout)), err
}

func GetSystemdDaemon(daemonName string) (Daemon, error) {
	isActiveStr, activeErr := execSystemctl("is-active", daemonName)
	isEnabledStr, enabledErr := execSystemctl("is-enabled", daemonName)

	daemonInfo := Daemon{
		name:      daemonName,
		isActive:  getActivityIota(isActiveStr),
		isEnabled: getEnabledIota(isEnabledStr),
	}

	if activeErr != nil {
		return daemonInfo, activeErr
	}
	return daemonInfo, enabledErr
}

func GetDaemon(initSystem string, daemonName string) (Daemon, error) {
	if initSystem == systemd {
		return GetSystemdDaemon(daemonName)
	}
	return Daemon{}, errors.New("no known init system has been chosen")
}
