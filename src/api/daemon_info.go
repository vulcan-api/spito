package api

import (
	"errors"
	"os/exec"
	"strings"
)

const systemd = "systemctl"

type Daemon struct {
	name      string
	isActive  string
	isEnabled string
}

func execSystemctl(command string, daemon string) (string, error) {
	cmd := exec.Command(systemd, command, daemon)
	stdout, err := cmd.Output()
	return strings.TrimSpace(string(stdout)), err
}

func GetSystemdDaemon(daemonName string) (Daemon, error) {
	isActive, err := execSystemctl("is-active", daemonName)
	if err != nil {
		return Daemon{}, err
	}

	isEnabled, err := execSystemctl("is-enabled", daemonName)
	if err != nil {
		return Daemon{}, err
	}

	daemonInfo := Daemon{
		name:      daemonName,
		isActive:  isActive,
		isEnabled: isEnabled,
	}

	return daemonInfo, nil
}

func GetDaemon(daemonName string, initSystem string) (Daemon, error) {
	if initSystem == systemd {
		return GetSystemdDaemon(daemonName)
	}
	return Daemon{}, errors.New("no known init system has been chosen")
}
