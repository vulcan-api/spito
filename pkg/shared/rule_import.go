package shared

import (
	daemon_tracker "github.com/avorty/spito/pkg"
	"github.com/avorty/spito/pkg/package_conflict"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/godbus/dbus/v5"
)

type ImportLoopData struct {
	VRCT           vrct.RuleVRCT
	InfoApi        InfoInterface
	RulesHistory   RulesHistory
	ErrChan        chan error
	PackageTracker package_conflict.PackageConflictTracker
	Options        []string
	DaemonTracker  daemon_tracker.DaemonTracker
	DbusConn       *dbus.Conn
	GuiMode        bool
}

func (i *ImportLoopData) DeleteRuntimeTemp() error {
	return i.VRCT.DeleteRuntimeTemp()
}
