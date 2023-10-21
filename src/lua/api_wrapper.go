package lua

import (
	"github.com/Shopify/go-lua"
	"github.com/google/uuid"
	"github.com/nasz-elektryk/spito-rules/api"
	"github.com/oleiade/reflections"
	"reflect"
	"strings"
)

func registerApi(L *lua.State) {
	L.Register("getCurrentDistro", GetCurrentDistro)
	newGlobalConstructor(L, "Package", reflect.TypeOf(api.Package{}))
}

func newGlobalConstructor(L *lua.State, name string, Obj reflect.Type) {
	L.PushGoFunction(func(state *lua.State) int {
		userDataIdentifier := "_" + name + "_" + strings.ReplaceAll(uuid.NewString(), "-", "")
		obj := reflect.New(Obj)

		// TODO: think about some kind of deallocating this value
		L.PushUserData(obj)
		L.SetGlobal(userDataIdentifier)

		pushMetaTableStructInterface(L, userDataIdentifier)

		return 1
	})
	L.SetGlobal(name)
}

func pushMetaTableStructInterface(L *lua.State, userdataGlobalName string) {
	L.NewTable()
	L.PushString(userdataGlobalName)
	L.SetField(-2, "_userdata")

	L.NewTable()
	L.PushGoFunction(__index)
	L.SetField(-2, "__index")

	L.SetMetaTable(-2)
}

func __index(L *lua.State) int {
	key, ok := L.ToString(2)
	if !ok {
		return 0
	}

	L.RawGetValue(1, "_userdata")
	userdataName, ok := L.ToString(-1)
	if !ok {
		return 0
	}
	L.Pop(1)

	L.Global(userdataName)
	p := L.ToUserData(-1)
	L.Pop(1)

	field, err := reflections.GetField(p, key)
	if err == nil {
		err := addToStackBasedOnType(L, field, reflect.TypeOf(field).Kind())
		if err != nil {
			// TODO: consider better error handling
			return 0
		}
		return 1
	}

	// TODO: implement method handling

	return 0
}

func GetCurrentDistro(L *lua.State) int {
	distro := api.GetCurrentDistro()
	if err := AddStructToStack(L, distro); err != nil {
		panic(err)
	}

	return 1
}
