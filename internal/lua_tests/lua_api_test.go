package api_tests

import (
	"fmt"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	shared "github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
	"path/filepath"
	"testing"
)

type luaTest struct {
	file       string
	beforeTest func(t *testing.T) error
	afterTest  func(t *testing.T, revertNum int) error
}

const basePath = "/tmp/spito-lua-test/"
const exampleJsonName = "example.json"
const expectedExampleJsonContent = `{"first-key": "first-val", "example-key": "example-val", "next-example-key": "next-example-val"}`

var exampleJsonPath = filepath.Join(basePath, exampleJsonName)

func prepareFsTest(_ *testing.T) error {
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(exampleJsonPath, []byte(`{"first-key": "first-val"}`), 0755)
}

func finalizeFsTest(_ *testing.T, _ int) error {
	content, err := os.ReadFile(exampleJsonPath)
	if err != nil {
		return err
	}

	return vrctFs.CompareConfigs(content, []byte(expectedExampleJsonContent), vrctFs.JsonConfig)
}

func finalizeGitTest(_ *testing.T, _ int) error {
	return os.RemoveAll("/tmp/spito-test/nfdsa321980")
}

func finalizeRevertFuncTest(t *testing.T, revertNum int) error {
	revertSteps, err := vrctFs.NewRevertSteps()
	if err != nil {
		return err
	}

	if err := revertSteps.Deserialize(revertNum); err != nil {
		return err
	}

	err = revertSteps.Apply(checker.GetRevertRuleFn(cmdApi.InfoApi{}))
	if err != nil {
		return err
	}

	path := "/tmp/spito-test/2fr4738gh5132"
	exists, err := shared.PathExists(path)
	if err != nil {
		return err
	}

	if exists {
		_ = os.Remove(path)
		t.Fatalf("Revert function did not remove the `%s` file\n", path)
	}

	return nil
}

func TestLuaApi(t *testing.T) {
	scripts := []luaTest{
		{file: "daemon_test.lua"},
		{file: "fs_test.lua", beforeTest: prepareFsTest, afterTest: finalizeFsTest},
		{file: "package_test.lua"},
		{file: "rule_require_test.lua"},
		{file: "sh_test.lua"},
		{file: "sysinfo_test.lua"},
		{file: "git_test.lua", afterTest: finalizeGitTest},
		{file: "revertFunc.lua", afterTest: finalizeRevertFuncTest},
	}

	for _, script := range scripts {
		file, err := os.ReadFile(script.file)
		if err != nil {
			t.Fatal(err)
		}

		if script.beforeTest != nil {
			err = script.beforeTest(t)
			if err != nil {
				t.Fatalf("error occured during preparation stage of test '%s': %s", script.file, err)
			}
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

		doesRulePass, err := checker.CheckRuleScript(&runtimeData, string(file), "")
		if err != nil {
			t.Fatalf("Error occurred in script '%s' : %s", script.file, fmt.Sprint(err))
		}

		if !doesRulePass {
			logAndFail(t, "Rule %s did not pass!", script.file)
		}

		var ruleIdentifiers []vrctFs.Rule
		for _, rule := range runtimeData.RulesHistory {
			ruleIdentifiers = append(ruleIdentifiers, vrctFs.Rule{
				Url:          rule.Url,
				NameOrScript: rule.NameOrScript,
				IsScript:     rule.IsScript,
			})
		}

		revertNum, err := ruleVRCT.Apply(ruleIdentifiers)
		if err != nil {
			return
		}

		if err := ruleVRCT.DeleteRuntimeTemp(); err != nil {
			logAndFail(t, "Failed to remove temporary VRCT files: %s", err.Error())
		}

		if script.afterTest != nil {
			err = script.afterTest(t, revertNum)
			if err != nil {
				logAndFail(t, "error occurred during finalization stage of test '%s': %s", script.file, err)
			}
		}
	}
}

func logAndFail(t *testing.T, format string, args ...interface{}) {
	t.Logf(format, args...)
	t.Fail()
}
