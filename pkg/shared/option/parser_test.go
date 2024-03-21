package option

import (
	"reflect"
	"testing"
)

type getIndexOutsideCase struct {
	rawString      string
	wantedPosition int
	wantedError    bool
}

var getIndexOutsideCases = []getIndexOutsideCase{
	{`{params = {hair = blonde}}`, -1, false},
	{`{first = val},`, 13, false},
	{`{first = val} ,`, 14, false},
	{`{{,  ,,,},,{,,}},`, 16, false},
	{`{,  ,,,},,{,,}},`, 8, false},
	{``, -1, false},
	{`,`, 0, false},
	{`{,`, -1, true},
	{`},`, 1, false},
	{`{{}},`, 4, false},
	{`dog = { hairType?, age: int = 5 } age: int = 1 }`, -1, false},
}

func TestGetIndexOutside(t *testing.T) {
	for _, edgeCase := range getIndexOutsideCases {
		position, err := GetIndexOutside(edgeCase.rawString, "{", "}", ",")
		if edgeCase.wantedError {
			if err == nil {
				t.Fatal("expected error but everything is OK")
			}
			continue
		}

		if err != nil {
			t.Fatal(err)
		}

		if position != edgeCase.wantedPosition {
			t.Fatalf("received position (%d) doesn't match wanted one (%d) in test: \n'%s'\n", position, edgeCase.wantedPosition, edgeCase.rawString)

		}
	}
}

type appendOptionCase struct {
	raw         string
	desired     []Option
	wantedError bool
}

var appendOptionCases = []appendOptionCase{
	{"{enum?:{ONE;TWO;;}=ONE}", []Option{
		{"enum", "ONE", Enum, true, []string{"ONE", "TWO", "", ""}, nil},
	}, false},
	{"{enum?:{ONE;TWO}=ONE}", []Option{
		{"enum", "ONE", Enum, true, []string{"ONE", "TWO"}, nil},
	}, false},
	// TODO: find very good array
	{"{array?:list={1;2;3}}", []Option{
		{"array", []string{"1", "2", "3"}, List, true, nil, nil},
	}, false},
	// TODO: find a way to escape a colon
	{`{testString?:string=",testValue"}`, []Option{
		{"testString", ",testValue", String, true, nil, nil},
	}, true},
	{"{number?:int=0}", []Option{
		{"number", 0, Int, true, nil, nil},
	}, false},
	{"{sub={name=jarek,surname=goof},gender:{male;female}=male,sibling?:string}", []Option{
		{"sub", nil, Struct, false, nil, []Option{
			{"name", "jarek", Any, false, nil, nil},
			{"surname", "goof", Any, false, nil, nil}},
		},
		{"gender", "male", Enum, false, []string{"male", "female"}, nil},
		{"sibling", nil, String, true, nil, nil}}, false,
	},
	{`{ageO?:int=1,position="leader",positionO?=2,nameO?:string,lastnameO?,dog={hairType?,smallDog={exists:bool=false},age:int=5},last?}`, []Option{
		{"ageO", 1, Int, true, nil, nil},
		{"position", "leader", Any, false, nil, nil},
		{"positionO", 2, Any, true, nil, nil},
		{"nameO", nil, String, true, nil, nil},
		{"lastnameO", nil, Any, true, nil, nil},
		{"dog", nil, Struct, false, nil, []Option{
			{"hairType", nil, Any, true, nil, nil},
			{"smallDog", nil, Struct, false, nil, []Option{
				{"exists", false, Bool, false, nil, nil},
			}},
			{"age", 5, Int, false, nil, nil}},
		},
		{"last", nil, Any, true, nil, nil}}, false,
	},
}

func TestAppendOptions(t *testing.T) {
	for _, testCase := range appendOptionCases {
		var obtained []Option
		obtained, err := AppendOptions(obtained, testCase.raw)
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
