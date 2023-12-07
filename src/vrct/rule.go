package vrct

import "reflect"

func NewRuleVRCT() *RuleVRCT {
	return &RuleVRCT{
		Vrcts: make([]VRCT, 0),
	}
}

type RuleVRCT struct {
	Vrcts []VRCT
}

func (v RuleVRCT) Of(vrctType VRCT) VRCT {
	for _, vrct := range v.Vrcts {
		if reflect.TypeOf(vrct) == reflect.TypeOf(vrctType) {
			return vrct
		}
	}

	newVrct := reflect.New(reflect.TypeOf(vrctType).Elem()).Interface().(VRCT)

	v.Vrcts = append(v.Vrcts, newVrct)
	return newVrct
}

func (v RuleVRCT) InnerValidate() error {
	for _, vrct := range v.Vrcts {
		err := vrct.InnerValidate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (v RuleVRCT) EnsureInitialized() error {
	for _, vrct := range v.Vrcts {
		err := vrct.EnsureInitialized()
		if err != nil {
			return err
		}
	}

	return nil
}

func (v RuleVRCT) Apply() error {
	for _, vrct := range v.Vrcts {
		err := vrct.Apply()
		if err != nil {
			return err
		}
	}

	return nil
}
