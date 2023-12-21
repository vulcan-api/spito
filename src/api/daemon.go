package api

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

var (
	ErrRequiresRoot          = errors.New("operation requires root")
	ErrUnsupportedInit       = errors.New("unknown init system")
	ErrDaemonDoesNotExist    = errors.New("daemon does not exist")
	ErrUnknownDirectory      = errors.New("init system uses unknown init scripts directory")
	ErrUnsupportedInitOutput = errors.New("init system produced unknown output")
)

type Daemon struct {
	Name      string
	IsActive  bool
	IsEnabled bool
	RunLevel  string
}

func execute(ctx context.Context, name string, args ...string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(cmd.Env, "LC_ALL=C")
	rawOutput, err := cmd.Output()
	output := string(rawOutput)

	if _, ok := ctx.Deadline(); !ok {
		return output, false, ctx.Err()
	}

	clearOutput := strings.TrimSpace(output)

	return clearOutput, true, err
}

func getDaemonDataFromFS(daemonName, path string) (bool, string, error) {
	runLevels, err := os.ReadDir(path)
	if err != nil {
		return false, "", ErrUnknownDirectory
	}

	runLevel := ""
	for _, e := range runLevels {
		daemonsInRL, err := os.ReadDir(path + e.Name())
		if err != nil {
			return runLevel != "", runLevel, err
		}

		for _, daemonInRL := range daemonsInRL {
			// It may be useful to change daemon.RunLevel from string to []string
			if daemonInRL.Name() == daemonName {
				runLevel = e.Name()
			}
		}
	}
	return runLevel != "", runLevel, nil
}

func getSystemdDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	daemon := Daemon{Name: daemonName}

	output, ok, err := execute(ctx, "systemctl", "is-active", daemonName)
	if !ok {
		return daemon, err
	}
	if output == "active" {
		daemon.IsActive = true
	}

	output, ok, err = execute(ctx, "systemctl", "is-enabled", daemonName)
	if !ok {
		return daemon, err
	}
	if output == "enabled" || output == "static" || output == "indirect" {
		daemon.IsEnabled = true
	}

	output, ok, err = execute(ctx, "systemctl", "get-default")
	if !ok {
		return daemon, err
	}
	daemon.RunLevel = output

	return daemon, nil
}

func getRootlessOpenRCDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	daemon := Daemon{Name: daemonName}

	cmd := exec.CommandContext(ctx, "rc-status", "-f", "ini")
	rawOutput, err := cmd.Output()

	if _, ok := ctx.Deadline(); !ok {
		return daemon, ctx.Err()
	}

	if err != nil {
		return daemon, err
	}

	cfg, err := ini.Load(rawOutput)
	if err != nil {
		return daemon, err
	}
	runLevels := cfg.Sections()

	for _, level := range runLevels {
		if level.HasKey(daemonName) {
			daemon.RunLevel = level.Name()
			daemon.IsEnabled = true
			daemon.IsActive = false
			daemonInLevel := level.Key(daemonName)
			isStarted := daemonInLevel.Value()
			if isStarted == "started" {
				daemon.IsActive = true
			}
		}
	}

	return daemon, nil
}

func getOpenRCDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	daemon := Daemon{Name: daemonName}

	cmd := exec.CommandContext(ctx, "rc-service", daemonName, "status")
	rawOutput, err := cmd.Output()

	if _, ok := ctx.Deadline(); !ok {
		return daemon, ctx.Err()
	}

	if err != nil {
		return daemon, ErrDaemonDoesNotExist
	}

	output := string(rawOutput)

	if strings.Contains(output, "started") || strings.Contains(output, "is running") {
		daemon.IsActive = true
	} else if strings.Contains(output, "stopped") || strings.Contains(output, "is not running") {
		daemon.IsActive = false
	} else {
		return daemon, ErrUnsupportedInitOutput
	}

	cmd = exec.CommandContext(ctx, "rc-update", "-v", "show")
	rawOutput, err = cmd.Output()

	if _, ok := ctx.Deadline(); !ok {
		return daemon, ctx.Err()
	}

	if err != nil {
		return daemon, err
	}

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

	return daemon, nil
}

func getRunitDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	daemon := Daemon{Name: daemonName}

	if os.Geteuid() != 0 {
		return daemon, ErrRequiresRoot
	}

	output, ok, err := execute(ctx, "sv", "status", daemonName)

	if !ok {
		return daemon, err
	}

	if err != nil {
		return Daemon{}, ErrDaemonDoesNotExist
	}

	clearOut := strings.TrimSpace(output)

	daemon.IsActive = false

	okStatus := "run"

	statusLen := len(okStatus)
	realOutput := clearOut[0:statusLen]

	if realOutput == okStatus {
		daemon.IsActive = true
	}

	entries, err := os.ReadDir("/var/service")
	if err != nil {
		entries, err = os.ReadDir("/run/runit/service")
		if err != nil {
			return Daemon{}, ErrUnknownDirectory
		}
	}

	for _, e := range entries {
		if e.Name() == daemonName {
			daemon.IsEnabled = true
			break
		}
	}

	daemon.IsEnabled, daemon.RunLevel, err = getDaemonDataFromFS(daemonName, "/etc/runit/runsvdir/")
	if err != nil {
		return daemon, err
	}

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

	// TODO: root handling
	switch initSystem {
	case SYSTEMD:
		return getSystemdDaemon(ctx, daemonName)
	case OPENRC:
		return getOpenRCDaemon(ctx, daemonName)
	case RUNIT:
		return getRunitDaemon(ctx, daemonName)
	default:
		return Daemon{}, ErrUnsupportedInit
	}
}
