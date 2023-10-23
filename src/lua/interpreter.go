package lua

import (
	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
	"reflect"
)

const devMode = false

func DoesRulePasses(script string) (bool, error) {
	L := lua.NewState(lua.Options{SkipOpenLibs: !devMode})
	defer L.Close()
	attachApi(L)

	if err := L.DoString(script); err != nil {
		// TODO: think about better error handling!
		panic(err)
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
