package api_tests

import (
	"testing"
	"os"
	"github.com/nasz-elektryk/spito-rules/checker"
)

func TestGetInitSystem(t *testing.T) {

	script, err := os.ReadFile("sysinfo_test.lua")
	if err != nil {
		t.Fatal(err)
	}
	
	doesRulePass, err := checker.CheckRuleScript(string(script))
	if err != nil {
		t.Fatal(err)
	}
	
	println(doesRulePass)
}
