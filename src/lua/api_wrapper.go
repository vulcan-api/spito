package lua

import (
	"github.com/Shopify/go-lua"
	"github.com/nasz-elektryk/spito-rules/api"
)

func registerApi(L *lua.State) {
	L.Register("getCurrentDistro", GetCurrentDistro)
}

func GetCurrentDistro(L *lua.State) int {
	distro := api.GetCurrentDistro()
	if err := AddStructToStack(L, distro); err != nil {
		panic(err)
	}

	return 1
}
