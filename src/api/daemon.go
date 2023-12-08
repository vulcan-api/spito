package api

import (
	"context"
	"errors"
	"github.com/taigrr/systemctl"
	"os/exec"
	"strings"
	"time"
)

type Daemon struct {
	Name      string
	IsActive  bool
	IsEnabled bool
}

func getSystemdDaemon(ctx context.Context, daemonName string) (Daemon, error) {

	opts := systemctl.Options{UserMode: false}
	unit := daemonName

	isActive, err := systemctl.IsActive(ctx, unit, opts)
	if err != nil {
		return Daemon{}, err
	}

	isEnabled, err := systemctl.IsEnabled(ctx, unit, opts)
	if err != nil {
		return Daemon{}, err
	}

	return Daemon{
		Name:      daemonName,
		IsActive:  isActive,
		IsEnabled: isEnabled,
	}, nil
}

func getOpenRCDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	args := []string{daemonName, "status"}
	cmd := exec.CommandContext(ctx, "rc-service", args...)
	rawOutput, err := cmd.Output()
	if err != nil {
		return Daemon{}, err
	}
	output := strings.TrimSpace(string(rawOutput))
	rawStatus := strings.TrimLeft(output, "* status: ")
	status := strings.TrimSpace(string(rawStatus))

	daemon := Daemon{}

	switch status {
	case "started":
		daemon.IsActive = true
		break
	case "stopped":
		daemon.IsActive = false
		break
	default:
		return Daemon{}, errors.New("unknown output of rc-service")
	}
	return daemon, nil
}

func GetDaemon(daemonName string) (Daemon, error) {
	initSystem, err := GetInitSystem()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err != nil {
		return Daemon{}, err
	}
	switch initSystem {
	case SYSTEMD:
		return getSystemdDaemon(ctx, daemonName)
	case OPENRC:
		return Daemon{}, err
	case RUNIT:
		return Daemon{}, err
	default:
		return Daemon{}, errors.New("unknown init system")
	}
}
