package checker

import (
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"github.com/yuin/gopher-lua"
	"os"
	"path/filepath"
)

const rulesetDirConstantName = "RULESET_DIR"

func GetLuaState(script string, importLoopData *shared.ImportLoopData, ruleConf *shared.RuleConfigLayout, rulesetPath string) (*lua.LState, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})

	// Standard libraries
	lua.OpenString(L)

	L.SetGlobal(rulesetDirConstantName, lua.LString(rulesetPath))
	attachApi(importLoopData, ruleConf, L)
	attachRuleRequiring(importLoopData, L)

	return L, L.DoString(script)
}

func ExecuteLuaMain(L *lua.LState) (bool, error) {
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

func ExecuteLuaRevert(L *lua.LState) (bool, error) {
	revertFn := L.GetGlobal("revert")
	if revertFn.Type() == lua.LTNil {
		return true, nil
	}

	err := L.CallByParam(lua.P{
		Fn:      L.GetGlobal("revert"),
		Protect: true,
		NRet:    1,
	})
	if err != nil {
		return false, err
	}

	return bool(L.Get(-1).(lua.LBool)), nil
}

func attachRuleRequiring(importLoopData *shared.ImportLoopData, L *lua.LState) {
	L.SetGlobal("require_remote", L.NewFunction(func(state *lua.LState) int {
		rulesetIdentifier := L.Get(1).String()
		ruleName := L.Get(2).String()

		doesRulePass, err := CheckRuleByIdentifier(importLoopData, rulesetIdentifier, ruleName)
		handleErrorAndPanic(importLoopData.ErrChan, err)

		rulesetLocation := NewRulesetLocation(rulesetIdentifier, false)

		if err = L.DoFile(filepath.Join(rulesetLocation.GetRulesetPath(), "rules", ruleName+".lua")); err != nil {
			importLoopData.ErrChan <- err
			panic(nil)
		}

		if !doesRulePass {
			importLoopData.ErrChan <- fmt.Errorf("rule %s/%s did not pass requirements", rulesetIdentifier, ruleName)
			panic(nil)
		}
		return 0
	}))

	L.SetGlobal("require_file", L.NewFunction(func(state *lua.LState) int {
		rulePath := L.Get(1).String()

		err := shared.ExpandTilde(&rulePath)
		if err != nil {
			importLoopData.ErrChan <- err
			panic(nil)
		}

		script, err := os.ReadFile(rulePath)
		handleErrorAndPanic(importLoopData.ErrChan, err)

		if err = L.DoString(string(script)); err != nil {
			importLoopData.ErrChan <- err
			panic(nil)
		}

		doesRulePass, err := CheckRuleScript(importLoopData, string(script), filepath.Dir(rulePath))
		handleErrorAndPanic(importLoopData.ErrChan, err)

		if !doesRulePass {
			importLoopData.ErrChan <- fmt.Errorf("rule from %s did not pass requirements", rulePath)
			panic(nil)
		}

		return 0
	}))
}
