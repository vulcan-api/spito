package integration_tests

import (
	"github.com/nasz-elektryk/spito-rules/checker"
	"os"
	"testing"
)

func TestRuleRequire(t *testing.T) {
	script, err := os.ReadFile("./rule_require_test.lua")
	if err != nil {
		t.Fatal(err)
	}

	doesRulePasses, err := checker.CheckRuleScript(string(script))
	if err != nil {
		t.Fatal(err)
	}
	
	println("rule passes: ", doesRulePasses)
	if !doesRulePasses {
		t.Fatal("Rule failed")
	}
}