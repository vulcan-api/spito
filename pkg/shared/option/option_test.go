package option

import (
	"testing"
)

type setValueCase struct {
	testOption   Option
	desiredValue string
	wantedError  bool
}

var setValueCases = []setValueCase{
	{Option{"food", "susi", String, false, nil, nil}, "josh", false},
	{Option{"bool", "false", Bool, false, nil, nil}, "asd", true},
	{Option{"enum", "two", Enum, false, []string{"one", "two", "three"}, nil}, "three", false},
	{Option{"enum", "two", Enum, false, []string{"one", "two", "three"}, nil}, "four", true},
}

func TestOption_SetValue(t *testing.T) {
	for _, testCase := range setValueCases {
		err := testCase.testOption.SetValue(testCase.desiredValue)
		if testCase.wantedError && err != nil {
			continue
		}
		if err != nil {
			t.Fatal(err.Error())
		}
	}
}
