package shared

import (
	"github.com/avorty/spito/pkg/vrct"
)

type ImportLoopData struct {
	VRCT         vrct.RuleVRCT
	InfoApi      InfoInterface
	RulesHistory RulesHistory
	ErrChan      chan error
}
