package checker

import (
	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
	"reflect"
)

const devMode = false

func ExecuteLuaMain(script string, rulesHistory *RulesHistory) (bool, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: !devMode})
	defer L.Close()
	
	attachApi(L)
	attachRuleRequiring(L, rulesHistory)

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

func attachRuleRequiring(L *lua.LState, rulesHistory *RulesHistory) {
	L.SetGlobal("require_rule", L.NewFunction(func(state *lua.LState) int {
		ruleUrl := L.Get(1).String()
		ruleName := L.Get(2).String()
		
		println("arg2, ", ruleName)
		
		result := CheckRule(rulesHistory, ruleUrl, ruleName)
		L.Push(lua.LBool(result))

		return 1
	}))
}

func setGlobalConstructor(L *lua.LState, name string, Obj reflect.Type) {
	L.SetGlobal(name, L.NewFunction(func(state *lua.LState) int {
		obj := reflect.New(Obj)

		L.Push(luar.New(L, obj.Interface()))
		return 1
	}))
}

func setGlobalFunction(L *lua.LState, name string, fn interface{}) {
	L.SetGlobal(name, luar.New(L, fn))
}
