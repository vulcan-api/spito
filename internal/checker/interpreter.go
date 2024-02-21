package checker

import (
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/shared/option"
	"github.com/yuin/gopher-lua"
	"os"
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
	ruleOptionType := ruleOption.Type
	if ruleOption.Type == option.Struct {
		value = getOptions(ruleOption.Options, L)
	} else if ruleOption.DefaultValue == nil {
		value = lua.LNil
	} else {
		if ruleOptionType == option.Any {
			ruleOptionType = option.GetType(ruleOption.DefaultValue)
		}
		switch ruleOptionType {
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
			value = lua.LNil
		}
	}
	return value
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
