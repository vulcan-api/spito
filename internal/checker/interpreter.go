package checker

import (
	"fmt"
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

		rulesetLocation := NewRulesetLocation(rulesetIdentifier)
		if !rulesetLocation.IsRuleSetDownloaded() {
			err := FetchRuleset(&rulesetLocation)
			handleErrorAndPanic(importLoopData.ErrChan, err)
		}

	        err := L.DoFile(filepath.Join(rulesetLocation.GetRulesetPath(), "rules", fmt.Sprintf("%s.lua", ruleName)))
		handleErrorAndPanic(importLoopData.ErrChan, err)
		return 0
	}))

	L.SetGlobal("require_file", L.NewFunction(func(state *lua.LState) int {
		rulePath := L.Get(1).String()
		shared.ExpandTilde(&rulePath)
		
		if err := L.DoFile(rulePath); err != nil {
			importLoopData.ErrChan <- err
			panic(nil)
		}

		return 0
	}))
}
