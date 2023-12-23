package vrct

import (
	"github.com/nasz-elektryk/spito/pkg/vrct/vrctFs"
)

func NewRuleVRCT() (*RuleVRCT, error) {
	fsVRCT, err := vrctFs.NewFsVRCT()
	if err != nil {
		return nil, err
	}

	return &RuleVRCT{
		Fs: fsVRCT,
	}, nil
}

type RuleVRCT struct {
	Fs vrctFs.FsVRCT
}

func (v RuleVRCT) InnerValidate() error {
	return v.Fs.InnerValidate()
}

func (v RuleVRCT) Apply() error {
	return v.Fs.Apply()
}
