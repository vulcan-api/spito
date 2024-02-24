package shared

import (
	"github.com/avorty/spito/pkg/package_conflict"
	"github.com/avorty/spito/pkg/vrct"
)

type ImportLoopData struct {
	VRCT           vrct.RuleVRCT
	InfoApi        InfoInterface
	RulesHistory   RulesHistory
	ErrChan        chan error
	PackageTracker package_conflict.PackageConflictTracker
}

func (i *ImportLoopData) DeleteRuntimeTemp() error {
	return i.VRCT.DeleteRuntimeTemp()
}
