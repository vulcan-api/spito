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

	L.SetGlobal("_O", getOptions(ruleConf.Options, L))
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

func getOptions(options []option.Option, L *lua.LState) lua.LValue {
	structNamespace := newLuaNamespace()

	for _, ruleOption := range options {
		structNamespace.AddField(ruleOption.Name, getOptionLValue(ruleOption, L))
	}

	return structNamespace.createTable(L)
}

func getOptionLValue(ruleOption option.Option, L *lua.LState) lua.LValue {
	var value lua.LValue
	if ruleOption.Type == option.Struct {
		value = getOptions(ruleOption.Options, L)
	} else {
		value = getAnyLValue(ruleOption.DefaultValue, ruleOption.Type)
	}
	return value
}

func getAnyLValue(rawValue any, optionType option.Type) lua.LValue {
	value := lua.LNil
	realOptionType := optionType
	if rawValue != nil {
		if optionType == option.Any || optionType == option.Enum {
			realOptionType = option.GetType(rawValue)
		}
		switch realOptionType {
		case option.Int:
			value = lua.LNumber(rawValue.(int))
			break
		case option.UInt:
			value = lua.LNumber(rawValue.(uint))
			break
		case option.Float:
			value = lua.LNumber(rawValue.(float64))
			break
		case option.String:
			value = lua.LString(rawValue.(string))
			break
		case option.Bool:
			value = lua.LBool(rawValue.(bool))
			break
		default:
			value = lua.LNil
		}
	}
	return value
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
