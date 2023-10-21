package lua

import (
	"github.com/Shopify/go-lua"
)

func DoesRulePasses(luaScript string) (bool, error) {
	L := lua.NewState()
	registerApi(L)
	lua.OpenLibraries(L)

	if err := lua.DoString(L, luaScript); err != nil {
		return false, err
	}
	
	L.Global("main")
	L.Call(0, 1)

	return L.ToBoolean(-1), nil
}
