package checker

import (
	"fmt"
	"github.com/avorty/spito/pkg/path"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/shared/option"
	"github.com/yuin/gopher-lua"
	"os"
	"path/filepath"
	"strconv"
)

const rulesetDirConstantName = "RULESET_DIR"

func GetLuaState(script string, importLoopData *shared.ImportLoopData, ruleConf *shared.RuleConfigLayout, rulesetPath string) (*lua.LState, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})

	// Standard libraries
	lua.OpenString(L)
	lua.OpenBase(L)

	L.SetGlobal(rulesetDirConstantName, lua.LString(rulesetPath))

	options, err := option.Compare(importLoopData.Options, ruleConf.Options)
	if err != nil {
		return L, err
	}

	luaOptions, err := getOptions(options, L)
	if err != nil {
		return L, err
	}

	L.SetGlobal("OPTIONS", luaOptions)
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

func getOptions(options []option.Option, L *lua.LState) (lua.LValue, error) {
	structNamespace := newLuaNamespace()

	for _, ruleOption := range options {
		value, err := getOptionLValue(ruleOption, L)
		if err != nil {
			return structNamespace.createTable(L), err
		}

		structNamespace.AddField(ruleOption.Name, value)
	}

	return structNamespace.createTable(L), nil
}

func getOptionLValue(ruleOption option.Option, L *lua.LState) (lua.LValue, error) {
	var value lua.LValue
	var err error
	switch ruleOption.Type {
	case option.Struct:
		value, err = getOptions(ruleOption.Options, L)
		break
	case option.List:
		value, err = getLList(ruleOption, L)
		break
	default:
		value, err = getAnyLValue(ruleOption.DefaultValue, ruleOption.Type)
	}
	return value, err
}

// TODO: allow to access array element by int not string
func getLList(ruleOption option.Option, L *lua.LState) (lua.LValue, error) {
	list := newLuaNamespace()

	for i, value := range ruleOption.DefaultValue.([]string) {
		value, err := getAnyLValue(value, option.String)
		if err != nil {
			return list.createTable(L), err
		}

		list.AddField(strconv.Itoa(i), value)
	}

	return list.createTable(L), nil
}

func getAnyLValue(rawValue any, optionType option.Type) (lua.LValue, error) {
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
			return value, fmt.Errorf("value '%s' of type '%d' is not supported", rawValue, optionType)
		}
	}
	return value, nil
}

func attachRuleRequiring(importLoopData *shared.ImportLoopData, L *lua.LState) {
	L.SetGlobal("require_remote", L.NewFunction(func(state *lua.LState) int {
		rulesetIdentifier := L.Get(1).String()
		ruleName := L.Get(2).String()

		doesRulePass, err := CheckRuleByIdentifier(importLoopData, rulesetIdentifier, ruleName)
		handleErrorAndPanic(importLoopData.ErrChan, err)

		rulesetLocation, err := NewRulesetLocation(rulesetIdentifier, false)
		handleErrorAndPanic(importLoopData.ErrChan, err)

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

		err := path.ExpandTilde(&rulePath)
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
