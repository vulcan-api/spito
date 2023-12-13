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
	InitLevel string
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
		InitLevel: "", // TODO
	}, nil
}

func getOpenRCDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	cmd := exec.CommandContext(ctx, "rc-service", daemonName, "status")
	rawOutput, err := cmd.Output()
	if err != nil {
		return Daemon{}, err
	}

	output := string(rawOutput)

	daemon := Daemon{}

	if strings.Contains(output, "started") {
		daemon.IsActive = true
	} else if strings.Contains(output, "stopped") {
		daemon.IsActive = false
	} else {
		return Daemon{}, errors.New(output)
	}

	cmd = exec.CommandContext(ctx, "rc-update", "-v", "show")
	rawOutput, err = cmd.Output()

	if err != nil {
		return Daemon{}, err
	}

	daemonsAndInitLevels := strings.Split(string(rawOutput), "|")
	if len(daemonsAndInitLevels) > 1 {
		initLevel := strings.TrimSpace(daemonsAndInitLevels[1])
		if initLevel != "" {
			daemon.InitLevel = initLevel
			daemon.IsEnabled = true
		}
	}

	daemon.Name = daemonName

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
		return getOpenRCDaemon(ctx, daemonName)
	case RUNIT:
		return Daemon{}, err
	default:
		return Daemon{}, errors.New("unknown init system")
	}
}
