package option

import (
	"reflect"
	"testing"
)

type compareCase struct {
	userOptions      []string
	specifiedOptions []Option
	desired          []Option
	wantedError      bool
}

var compareCases = []compareCase{
	{
		[]string{"food=good", "name=better"},
		[]Option{
			{"food", "susi", String, false, nil, nil},
			{"name", "josh", String, false, nil, nil},
			{"age", 4, Int, false, nil, nil},
		},
		[]Option{
			{"food", "good", String, false, nil, nil},
			{"name", "better", String, false, nil, nil},
			{"age", 4, Int, false, nil, nil},
		},
		false,
	},
}

func TestCompare(t *testing.T) {
	for _, testCase := range compareCases {
		var obtained []Option
		obtained, err := Compare(testCase.userOptions, testCase.specifiedOptions)
		if testCase.wantedError && err != nil {
			continue
		}
		if err != nil {
			t.Fatal(err.Error())
		}
		if !reflect.DeepEqual(testCase.desired, obtained) {
			t.Fatalf("desired \n>>>\n%#v\n<<< and obtained \n>>>\n%#v\n<<<", testCase.desired, obtained)
		}
	}
}
