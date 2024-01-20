package vrct

import (
	"github.com/avorty/spito/pkg/vrct/vrctFs"
)

type RuleVRCT struct {
	Fs vrctFs.VRCTFs
}

func NewRuleVRCT() (*RuleVRCT, error) {
	fsVRCT, err := vrctFs.NewFsVRCT()
	if err != nil {
		return nil, err
	}

	return &RuleVRCT{
		Fs: fsVRCT,
	}, nil
}

func (v RuleVRCT) DeleteRuntimeTemp() error {
	return v.Fs.DeleteRuntimeTemp()
}

func (v RuleVRCT) Apply() (int, error) {
	return v.Fs.Apply()
}

func (v RuleVRCT) Revert() error {
	return v.Revert()
}
