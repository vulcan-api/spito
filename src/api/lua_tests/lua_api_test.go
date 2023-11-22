package api_tests

import (
	"github.com/nasz-elektryk/spito/checker"
	"github.com/nasz-elektryk/spito/cmd/cmdApi"
	"os"
	"testing"
)

func TestLuaApi(t *testing.T) {

	scripts := []string{
		"sysinfo_test.lua",
		//"fs_test.lua",
		"rule_require_test.lua",
		"daemon_test.lua",
		"package_test.lua",
	}

	for _, scriptName := range scripts {
		file, err := os.ReadFile(scriptName)
		if err != nil {
			t.Fatal(err)
		}

		runtimeData := checker.RuntimeData{
			RulesHistory: checker.RulesHistory{},
			ErrChan:      make(chan error),
			InfoApi:      cmdApi.InfoApi{},
		}

		doesRulePass, err := checker.CheckRuleScript(&runtimeData, string(file))
		if err != nil {
			t.Fatal(err)
		}

		if !doesRulePass {
			t.Fatalf("Rule %s did not pass!", scriptName)
		}
	}
}
