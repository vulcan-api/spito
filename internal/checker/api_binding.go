package checker

import (
	api "github.com/avorty/spito/pkg/api"
	"github.com/avorty/spito/pkg/shared"
	"github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"reflect"
)

// Every cmdApi needs to be attached here in order to be available:
func attachApi(importLoopData *shared.ImportLoopData, ruleConf *RuleConf, L *lua.LState) {
	apiNamespace := newLuaNamespace()

	apiNamespace.AddField("pkg", getPackageNamespace(L))
	apiNamespace.AddField("sys", getSysInfoNamespace(L))
	apiNamespace.AddField("fs", getFsNamespace(L, importLoopData))
	apiNamespace.AddField("info", getInfoNamespace(importLoopData, L))

	if ruleConf.Unsafe {
		apiNamespace.AddField("sh", getShNamespace(L))
	}

	apiNamespace.setGlobal(L, "api")
}

func getPackageNamespace(L *lua.LState) lua.LValue {
	pkgNamespace := newLuaNamespace()
	pkgNamespace.AddFn("Get", api.GetPackage)

	return pkgNamespace.createTable(L)
}

func getSysInfoNamespace(L *lua.LState) lua.LValue {
	sysInfoNamespace := newLuaNamespace()

	sysInfoNamespace.AddFn("GetDistro", api.GetDistro)
	sysInfoNamespace.AddFn("GetDaemon", api.GetDaemon)
	sysInfoNamespace.AddFn("GetInitSystem", api.GetInitSystem)

	return sysInfoNamespace.createTable(L)
}

func getFsNamespace(L *lua.LState, importLoop *shared.ImportLoopData) lua.LValue {
	fsNamespace := newLuaNamespace()

	apiFs := api.FsApi{
		FsVRCT: &importLoop.VRCT.Fs,
	}

	fsNamespace.AddFn("PathExists", apiFs.PathExists)
	fsNamespace.AddFn("FileExists", apiFs.FileExists)
	fsNamespace.AddFn("ReadFile", apiFs.ReadFile)
	fsNamespace.AddFn("ReadDir", apiFs.ReadDir)
	fsNamespace.AddFn("FileContains", apiFs.FileContains)
	fsNamespace.AddFn("RemoveComments", apiFs.RemoveComments)
	fsNamespace.AddFn("Find", apiFs.Find)
	fsNamespace.AddFn("FindAll", apiFs.FindAll)
	fsNamespace.AddFn("GetProperLines", apiFs.GetProperLines)
	fsNamespace.AddFn("CreateFile", apiFs.CreateFile)

	return fsNamespace.createTable(L)
}

func getInfoNamespace(importLoopData *shared.ImportLoopData, L *lua.LState) lua.LValue {
	infoApi := importLoopData.InfoApi
	infoNamespace := newLuaNamespace()

	infoNamespace.AddFn("Log", infoApi.Log)
	infoNamespace.AddFn("Debug", infoApi.Debug)
	infoNamespace.AddFn("Error", infoApi.Error)
	infoNamespace.AddFn("Warn", infoApi.Warn)
	infoNamespace.AddFn("Important", infoApi.Important)

	return infoNamespace.createTable(L)
}
func getShNamespace(L *lua.LState) lua.LValue {
	shellNamespace := newLuaNamespace()

	// TODO: add shell functions here

	return shellNamespace.createTable(L)
}

type LuaNamespace struct {
	constructors map[string]reflect.Type
	functions    map[string]interface{}
	fields       map[string]lua.LValue
}

func newLuaNamespace() LuaNamespace {
	return LuaNamespace{
		constructors: map[string]reflect.Type{},
		functions:    make(map[string]interface{}),
		fields:       make(map[string]lua.LValue),
	}
}

func (ln LuaNamespace) AddConstructor(name string, Obj reflect.Type) {
	ln.constructors[name] = Obj
}

func (ln LuaNamespace) AddFn(name string, fn interface{}) {
	ln.functions[name] = fn
}

func (ln LuaNamespace) AddField(name string, field lua.LValue) {
	ln.fields[name] = field
}

func (ln LuaNamespace) setGlobal(L *lua.LState, name string) {
	namespaceTable := ln.createTable(L)
	L.SetGlobal(name, namespaceTable)
}

func (ln LuaNamespace) createTable(L *lua.LState) *lua.LTable {
	namespaceTable := L.NewTable()

	for fnName, fn := range ln.functions {
		L.SetField(namespaceTable, fnName, luar.New(L, fn))
	}
	for constrName, constrInterface := range ln.constructors {
		constr := constructorFunction(L, constrInterface)
		L.SetField(namespaceTable, constrName, constr)
	}
	for fieldName, field := range ln.fields {
		L.SetField(namespaceTable, fieldName, field)
	}

	return namespaceTable
}

func constructorFunction(L *lua.LState, Obj reflect.Type) lua.LValue {
	return L.NewFunction(func(state *lua.LState) int {
		obj := reflect.New(Obj)

		state.Push(luar.New(state, obj.Interface()))
		return 1
	})
}
