package checker

import (
	"github.com/yuin/gopher-lua"
)

const devMode = true

func ExecuteLuaMain(script string, rulesHistory *RulesHistory, errChan chan error) (bool, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: !devMode})
	defer L.Close()

	attachApi(L)
	attachRuleRequiring(L, rulesHistory, errChan)

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

func attachRuleRequiring(L *lua.LState, rulesHistory *RulesHistory, errChan chan error) {
	L.SetGlobal("require_rule", L.NewFunction(func(state *lua.LState) int {
		ruleUrl := L.Get(1).String()
		ruleName := L.Get(2).String()

		result := _internalCheckRule(rulesHistory, errChan, ruleUrl, ruleName)
		L.Push(lua.LBool(result))

		return 1
	}))
}
