package api

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/taigrr/systemctl"
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

func getDaemonDataFromFS(daemonName, path string) (bool, string, error) {

	runlevels, err := os.ReadDir(path)
	if err != nil {
		return false, "", ErrUnknownDirectory
	}

	runLevel := ""
	for _, e := range runlevels {
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

// TODO: amke it library independent
func getSystemdDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	daemon := Daemon{}

	opts := systemctl.Options{UserMode: false}
	unit := daemonName

	IsActive, err := systemctl.IsActive(ctx, unit, opts)
	if err != nil {
		return daemon, err
	}
	daemon.IsActive = IsActive

	isEnabled, err := systemctl.IsEnabled(ctx, unit, opts)
	if err != nil {
		return daemon, err
	}
	daemon.IsEnabled = isEnabled

	return daemon, nil
}

func getRootlessOpenRCDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	daemon := Daemon{}
	daemon.Name = daemonName

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

// TODO: check it again and add error handling
func getOpenRCDaemon(ctx context.Context, daemonName string) (Daemon, error) {
	daemon := Daemon{}
	daemon.Name = daemonName

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
	daemon := Daemon{}
	daemon.Name = daemonName

	if os.Geteuid() != 0 {
		return daemon, ErrRequiresRoot
	}

	cmd := exec.CommandContext(ctx, "sv", "status", daemonName)
	rawOutput, err := cmd.Output()

	if _, ok := ctx.Deadline(); !ok {
		return daemon, ctx.Err()
	}

	if err != nil {
		return Daemon{}, ErrDaemonDoesNotExist
	}

	output := string(rawOutput)
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
		entries, err = os.ReadDir("/etc/service")
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
