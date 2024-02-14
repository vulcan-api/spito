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
	beforeTest func() error
	afterTest  func() error
}

const basePath = "/tmp/spito-lua-test/"
const exampleJsonName = "example.json"
const expectedExampleJsonContent = `{"first-key": "first-val", "example-key": "example-val", "next-example-key": "next-example-val"}`

var exampleJsonPath = filepath.Join(basePath, exampleJsonName)

func prepareFsTest() error {
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(exampleJsonPath, []byte(`{"first-key": "first-val"}`), 0755)
}

func finalizeFsTest() error {
	content, err := os.ReadFile(exampleJsonPath)
	if err != nil {
		return err
	}

	return vrctFs.CompareConfigs(content, []byte(expectedExampleJsonContent), vrctFs.JsonConfig)
}

func TestLuaApi(t *testing.T) {
	scripts := []luaTest{
		{file: "daemon_test.lua"},
		{file: "fs_test.lua", beforeTest: prepareFsTest, afterTest: finalizeFsTest},
		{file: "package_test.lua"},
		{file: "rule_require_test.lua"},
		{file: "sh_test.lua"},
		{file: "sysinfo_test.lua"},
	}

	for _, script := range scripts {
		file, err := os.ReadFile(script.file)
		if err != nil {
			t.Fatal(err)
		}

		if script.beforeTest != nil {
			err = script.beforeTest()
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
			t.Fatalf("Error occurred: %s", fmt.Sprint(err))
		}

		if !doesRulePass {
			t.Fatalf("Rule %s did not pass!", script.file)
		}

		_, err = ruleVRCT.Apply()
		if err != nil {
			return
		}

		if err := ruleVRCT.DeleteRuntimeTemp(); err != nil {
			t.Fatal("Failed to remove temporary VRCT files", err.Error())
		}

		if script.afterTest != nil {
			err = script.afterTest()
			if err != nil {
				t.Fatalf("error occured during finalization stage of test '%s': %s", script.file, err)
			}
		}
	}
}
