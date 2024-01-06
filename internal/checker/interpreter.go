package checker

import (
	"os"
	"path/filepath"

	"github.com/avorty/spito/pkg/shared"
	"github.com/yuin/gopher-lua"
)

const rulesetDirConstantName = "RULESET_DIR"

func ExecuteLuaMain(script string, importLoopData *shared.ImportLoopData, ruleConf *RuleConf, rulesetPath string) (bool, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	// Standard libraries
	lua.OpenString(L)

	L.SetGlobal(rulesetDirConstantName, lua.LString(rulesetPath))
	attachApi(importLoopData, ruleConf, L)
	attachRuleRequiring(importLoopData, ruleConf, L)

	if err := L.DoString(script); err != nil {
		return false, err
	}

	err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("main"),
		Protect: true,
		NRet:    1,
	})
	if err != nil {
		return false, err
	}

	return bool(L.Get(-1).(lua.LBool)), nil
}

func attachRuleRequiring(importLoopData *shared.ImportLoopData, ruleConf *RuleConf, L *lua.LState) {
	L.SetGlobal("require_remote", L.NewFunction(func(state *lua.LState) int {
		rulesetIdentifier := L.Get(1).String()
		ruleName := L.Get(2).String()

		result := _internalCheckRule(importLoopData, rulesetIdentifier, ruleName, ruleConf)
		L.Push(lua.LBool(result))

		return 1
	}))

	L.SetGlobal("require_file", L.NewFunction(func(state *lua.LState) int {
		rulesetPath := L.Get(1).String()

		scriptContents, err := os.ReadFile(rulesetPath)
		if err != nil {
			importLoopData.InfoApi.Error(err.Error())
			os.Exit(1)
		}

		rulesetPath, err = filepath.Abs(rulesetPath)
		if err != nil {
			importLoopData.InfoApi.Error("Cannot convert file path to the absolute path!")
			os.Exit(1)
		}

		doesRulePass, _ := CheckRuleScript(importLoopData, string(scriptContents), filepath.Dir(rulesetPath))
		L.Push(lua.LBool(doesRulePass))

		return 1
	}))
}
