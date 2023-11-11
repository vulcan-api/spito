package integration_tests

import (
	"github.com/nasz-elektryk/spito/checker"
	"os"
	"testing"
)

func TestRuleRequire(t *testing.T) {
	script, err := os.ReadFile("./rule_require_test.lua")
	if err != nil {
		t.Fatal(err)
	}

	doesRulePass, err := checker.CheckRuleScript(string(script))
	if err != nil {
		t.Fatal(err)
	}

	println("Rule passes: ", doesRulePass)
	if !doesRulePass {
		t.Fatal("Rule failed")
	}
}
