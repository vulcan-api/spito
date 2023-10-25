package tests

import (
	"github.com/nasz-elektryk/spito-rules/checker"
	"os"
	"testing"
)

func TestRuleRequire(t *testing.T) {
	script, err := os.ReadFile("./rule_require_test.lua")
	if err != nil {
		t.Error(err)
	}

	doesRulePasses, err := checker.CheckRuleScript(string(script))
	if err != nil {
		t.Error(err)
	}
	
	println("rule passes: ", doesRulePasses)
	if !doesRulePasses {
		t.Error("Rule failed")
	}
}