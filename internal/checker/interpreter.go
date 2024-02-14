package checker

import (
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/shared/option"
	"github.com/yuin/gopher-lua"
	"path/filepath"
)

const rulesetDirConstantName = "RULESET_DIR"

func ExecuteLuaMain(script string, importLoopData *shared.ImportLoopData, ruleConf *shared.RuleConfigLayout, rulesetPath string) (bool, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()

	// Standard libraries
	lua.OpenString(L)

	L.SetGlobal(rulesetDirConstantName, lua.LString(rulesetPath))
	//fmt.Printf("%+v", *ruleConf)

	attachOptions(ruleConf, L)
	attachApi(importLoopData, ruleConf, L)
	attachRuleRequiring(importLoopData, L)

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

func attachOptions(ruleConf *shared.RuleConfigLayout, L *lua.LState) {
	for _, ruleOption := range ruleConf.Options {
		var value lua.LValue
		switch ruleOption.Type {
		case option.Int:
			value = lua.LNumber(ruleOption.DefaultValue.(int))
			break
		case option.UInt:
			value = lua.LNumber(ruleOption.DefaultValue.(uint))
			break
		case option.Float:
			value = lua.LNumber(ruleOption.DefaultValue.(float64))
			break
		case option.String:
			value = lua.LString(ruleOption.DefaultValue.(string))
			break
		case option.Bool:
			value = lua.LBool(ruleOption.DefaultValue.(bool))
			break
		default:
			// TODO: handle any
			break
		}
		L.SetGlobal(ruleOption.Name, value)
	}
}

func attachRuleRequiring(importLoopData *shared.ImportLoopData, L *lua.LState) {
	L.SetGlobal("require_remote", L.NewFunction(func(state *lua.LState) int {
		rulesetIdentifier := L.Get(1).String()
		ruleName := L.Get(2).String()

		rulesetLocation := NewRulesetLocation(rulesetIdentifier, false)
		err := FetchRuleset(&rulesetLocation)
		handleErrorAndPanic(importLoopData.ErrChan, err)

		err = L.DoFile(filepath.Join(rulesetLocation.GetRulesetPath(), "rules", fmt.Sprintf("%s.lua", ruleName)))
		handleErrorAndPanic(importLoopData.ErrChan, err)
		return 0
	}))

	L.SetGlobal("require_file", L.NewFunction(func(state *lua.LState) int {
		rulePath := L.Get(1).String()
		err := shared.ExpandTilde(&rulePath)
		if err != nil {
			importLoopData.ErrChan <- err
			panic(nil)
		}

		if err := L.DoFile(rulePath); err != nil {
			importLoopData.ErrChan <- err
			panic(nil)
		}

		return 0
	}))
}
