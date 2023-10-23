package lua

import (
	"fmt"
	"github.com/Shopify/go-lua"
	"github.com/nasz-elektryk/spito-rules/api"
	"reflect"
)

func registerApi(L *lua.State) {
	newGlobalConstructor(L, "Package", reflect.TypeOf(api.Package{}))
}

func newGlobalConstructor(L *lua.State, name string, Obj reflect.Type) {
	L.PushGoFunction(func(state *lua.State) int {
		obj := reflect.New(Obj).Elem()
		obj.FieldByName("Name").SetString("test")
		pushMetaTableStruct(L, obj)

		return 1
	})
	L.SetGlobal(name)
}

func pushMetaTableStruct(L *lua.State, obj interface{}) {
	L.NewTable()
	L.PushUserData(obj)
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
	userdata := L.ToValue(-1).(reflect.Value)
	L.Pop(1)

	if field := userdata.FieldByName(key); field.IsValid() {
		err := addToStackBasedOnType(L, field.Interface(), field.Kind())
		if err != nil {
			return 0
		}
		return 1
	}

	isMethodDeclared := false
	var method reflect.Method

	// TODO: implement method handling
	if m, ok := userdata.Type().MethodByName(key); ok && m.IsExported() {
		isMethodDeclared = true
		method = m
	}
	// TODO: implement method handling v2
	if m, ok := reflect.PointerTo(userdata.Type()).MethodByName(key); ok && m.IsExported() {
		isMethodDeclared = true
		method = m
	}
	if !isMethodDeclared {
		return 0
	}
	pushMethodToStack(L, &userdata, method, func() {
		fmt.Printf("%+v\n\n\n", userdata)
		println(L.TypeOf(1).String()) // Tutaj mamy errora
		
		L.PushUserData(userdata)
		L.SetField(1, "_userdata")
	})
	
	return 1
}
