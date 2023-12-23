package shared

import (
	"github.com/nasz-elektryk/spito/pkg/vrct"
)

type ImportLoopData struct {
	VRCT         vrct.RuleVRCT
	InfoApi      InfoInterface
	RulesHistory RulesHistory
	ErrChan      chan error
}
