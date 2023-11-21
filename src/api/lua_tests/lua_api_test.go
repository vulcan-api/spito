package api_tests

import (
	"fmt"
	"github.com/nasz-elektryk/spito/checker"
	"os"
	"testing"
)

func TestLuaApi(t *testing.T) {

	scripts := []string{
		"sysinfo_test.lua",
		"fs_test.lua",
		"rule_require_test.lua",
		"daemon_test.lua",
		"package_test.lua",
	}

	for _, scriptName := range scripts {
		file, err := os.ReadFile(scriptName)
		if err != nil {
			t.Fatal(err)
		}

		doesRulePass, err := checker.CheckRuleScript(string(file))
		if err != nil {
			t.Fatalf("Error occurred: %s", fmt.Sprint(err))
		}

		if !doesRulePass {
			t.Fatalf("Rule %s did not pass!", scriptName)
		}
	}
}
