package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/taigrr/systemctl"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Daemon struct {
	Name      string
	IsActive  bool
	IsEnabled bool
	RunLevel  string
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
		RunLevel:  "", // TODO
	}, nil
}

func getOpenRCDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	cmd := exec.CommandContext(ctx, "rc-service", daemonName, "status")
	rawOutput, _ := cmd.Output()

	output := string(rawOutput)

	daemon := Daemon{}

	if strings.Contains(output, "started") || strings.Contains(output, "is running"){
		daemon.IsActive = true
	} else if strings.Contains(output, "stopped") || strings.Contains(output, "is not running"){
		daemon.IsActive = false
	} else {
		return Daemon{}, errors.New(output)
	}

	cmd = exec.CommandContext(ctx, "rc-update", "-v", "show")
	rawOutput, _ = cmd.Output()

	lines := strings.Split(string(rawOutput), "\n")

	daemon.IsEnabled = false

	for _, line := range lines {
		splitLine := strings.Split(string(line), "|")
		if len(splitLine) <= 1 {
			continue
		}

		splitLine[0] = strings.TrimSpace(splitLine[0])
		if splitLine[0] == daemonName {
			initLevel := strings.TrimSpace(splitLine[1])
			if initLevel != "" {
				daemon.RunLevel = initLevel
				daemon.IsEnabled = true
			}
		}
	}

	daemon.Name = daemonName

	return daemon, nil
}

func getRunitDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	cmd := exec.CommandContext(ctx, "sv", "status", daemonName)
	rawOutput, err := cmd.Output()
	if err != nil {
		return Daemon{}, errors.New(fmt.Sprintf("Runit service problably doesn't exist. Runit output: %s", err))
	}

	output := string(rawOutput)
	clearOut := strings.TrimSpace(output)

	daemon := Daemon{}
	daemon.IsActive = false

	okStatus := "run"

	statusLen := len(okStatus)
	realOutput := clearOut[0:statusLen]

	if realOutput == okStatus {
		daemon.IsActive = true
	}

	entries, err := os.ReadDir("/var/service")
	if err != nil {
		entries, err = os.ReadDir("/etc/service")
		if err != nil {
			return Daemon{}, err
		}
	}

	for _, e := range entries {
		if e.Name() == daemonName {
			daemon.IsEnabled = true
			break
		}
	}

	const runsvdir = "/etc/runit/runsvdir/"

	runlevels, err := os.ReadDir(runsvdir)
	if err != nil {
		return Daemon{}, err
	}

	for _, e := range runlevels {
		daemonsInRL, err := os.ReadDir(runsvdir + e.Name())
		if err != nil {
			return Daemon{}, err
		}

		for _, daemonInRL := range daemonsInRL {
			// It may be useful to change daemon.RunLevel from string to []string
			if daemonInRL.Name() == daemonName {
				daemon.RunLevel = e.Name()
			}
		}
	}

	daemon.Name = daemonName

	return daemon, nil
}

func GetDaemon(daemonName string) (Daemon, error) {
	initSystem, err := GetInitSystem()
	if err != nil {
		return Daemon{}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	daemonName = strings.TrimSpace(daemonName)
	isSafe, _ := regexp.MatchString(`^[A-Za-z0-9_]+$`, daemonName)

	if !isSafe {
		return Daemon{}, errors.New("daemon name contains illegal character")
	}

	switch initSystem {
	case SYSTEMD:
		return getSystemdDaemon(ctx, daemonName)
	case OPENRC:
		return getOpenRCDaemon(ctx, daemonName)
	case RUNIT:
		return getRunitDaemon(ctx, daemonName)
	default:
		return Daemon{}, errors.New("unknown init system")
	}
}
