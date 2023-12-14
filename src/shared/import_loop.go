package shared

import (
	"github.com/nasz-elektryk/spito/vrct"
)

type ImportLoopData struct {
	VRCT         vrct.RuleVRCT
	InfoApi      InfoInterface
	RulesHistory RulesHistory
	ErrChan      chan error
}
