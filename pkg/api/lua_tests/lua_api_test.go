package api_tests

import (
	"fmt"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	shared "github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"os"
	"testing"
)

func TestLuaApi(t *testing.T) {
	scripts := []string{
		"daemon_test.lua",
		"fs_test.lua",
		"package_test.lua",
		"rule_require_test.lua",
		"sh_test.lua",
		"sysinfo_test.lua",
	}

	for _, scriptName := range scripts {
		file, err := os.ReadFile(scriptName)
		if err != nil {
			t.Fatal(err)
		}

		ruleVRCT, err := vrct.NewRuleVRCT()
		if err != nil {
			t.Fatal("Failed to initialized rule VRCT", err)
		}

		runtimeData := shared.ImportLoopData{
			VRCT:         *ruleVRCT,
			RulesHistory: shared.RulesHistory{},
			ErrChan:      make(chan error),
			InfoApi:      cmdApi.InfoApi{},
		}

		doesRulePass, err := checker.CheckRuleScript(&runtimeData, string(file))
		if err != nil {
			t.Fatalf("Error occurred: %s", fmt.Sprint(err))
		}

		if !doesRulePass {
			t.Fatalf("Rule %s did not pass!", scriptName)
		}
	}
}
