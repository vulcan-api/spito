package lua

import (
	"github.com/nasz-elektryk/spito-rules/api"
	"github.com/yuin/gopher-lua"
	"reflect"
)

// Every api needs to be attached here in order to be available:
func attachApi(L *lua.LState) {
	var t = reflect.TypeOf
	
	setGlobalConstructor(L, "Package", t(api.Package{}))
	setGlobalFunction(L, "GetDistro", api.GetDistro)
	setGlobalFunction(L, "GetDaemon", api.GetDaemon)
}
