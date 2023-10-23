package api

import (
	"context"
	"github.com/taigrr/systemctl"
	"time"
)

type Daemon struct {
	name      string
	isActive  bool
	isEnabled bool
}

func getSystemdDaemon(daemonName string) (Daemon, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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
		name:      daemonName,
		isActive:  isActive,
		isEnabled: isEnabled,
	}, err
}

func GetDaemon(daemonName string) (Daemon, error) {
	return getSystemdDaemon(daemonName)
}
