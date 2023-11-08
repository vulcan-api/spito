package integration_tests

import (
	"github.com/nasz-elektryk/spito-rules/checker"
	"os"
	"testing"
)

func TestRules(t *testing.T) {
	testRules := []string{
		"./fs_rules_test.lua",
		"./rule_require_test.lua",
	}

	for _, ruleFile := range testRules {
		script, err := os.ReadFile(ruleFile)
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
}
