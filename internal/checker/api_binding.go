package checker

import (
	"github.com/avorty/spito/pkg/api"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
	"reflect"
)

// Every cmdApi needs to be attached here in order to be available:
func attachApi(importLoopData *shared.ImportLoopData, ruleConf *shared.RuleConfigLayout, L *lua.LState) {
	apiNamespace := newLuaNamespace()

	apiNamespace.AddField("pkg", getPackageNamespace(importLoopData, L))
	apiNamespace.AddField("sys", getSysInfoNamespace(L))
	apiNamespace.AddField("fs", getFsNamespace(L, importLoopData))
	apiNamespace.AddField("info", getInfoNamespace(importLoopData, L))

	if ruleConf.Unsafe {
		apiNamespace.AddField("sh", getShNamespace(L))
	}

	apiNamespace.setGlobal(L, "api")
}

func getPackageNamespace(importLoopData *shared.ImportLoopData, L *lua.LState) lua.LValue {
	pkgNamespace := newLuaNamespace()
	pkgNamespace.AddFn("get", api.GetPackage)
	pkgNamespace.AddFn("install", func(packageName string) error {
		err := importLoopData.PackageTracker.AddPackage(packageName)
		return err
	})
	pkgNamespace.AddFn("remove", func(packageName string) error {
		err := importLoopData.PackageTracker.RemovePackage(packageName)
		return err
	})

	return pkgNamespace.createTable(L)
}

func getSysInfoNamespace(L *lua.LState) lua.LValue {
	sysInfoNamespace := newLuaNamespace()

	sysInfoNamespace.AddFn("getDistro", api.GetDistro)
	sysInfoNamespace.AddFn("getDaemon", api.GetDaemon)
	sysInfoNamespace.AddFn("getInitSystem", api.GetInitSystem)

	return sysInfoNamespace.createTable(L)
}

func getFsNamespace(L *lua.LState, importLoop *shared.ImportLoopData) lua.LValue {
	fsNamespace := newLuaNamespace()

	apiFs := api.FsApi{
		FsVRCT: &importLoop.VRCT.Fs,
	}

	fsNamespace.AddFn("pathExists", apiFs.PathExists)
	fsNamespace.AddFn("fileExists", apiFs.FileExists)
	fsNamespace.AddFn("readFile", apiFs.ReadFile)
	fsNamespace.AddFn("readDir", apiFs.ReadDir)
	fsNamespace.AddFn("fileContains", apiFs.FileContains)
	fsNamespace.AddFn("removeComments", api.RemoveComments)
	fsNamespace.AddFn("find", apiFs.Find)
	fsNamespace.AddFn("findAll", apiFs.FindAll)
	fsNamespace.AddFn("getProperLines", apiFs.GetProperLines)
	fsNamespace.AddFn("createFile", apiFs.CreateFile)
	fsNamespace.AddFn("createConfig", apiFs.CreateConfig)
	fsNamespace.AddFn("updateConfig", apiFs.UpdateConfig)
	fsNamespace.AddFn("compareConfigs", apiFs.CompareConfigs)
	fsNamespace.AddFn("apply", apiFs.Apply)
	fsNamespace.AddField("config", getConfigEnums(L))

	return fsNamespace.createTable(L)
}

func getConfigEnums(L *lua.LState) lua.LValue {
	infoNamespace := newLuaNamespace()

	infoNamespace.AddField("json", lua.LNumber(vrctFs.JsonConfig))
	infoNamespace.AddField("yaml", lua.LNumber(vrctFs.YamlConfig))
	infoNamespace.AddFn("toml", lua.LNumber(vrctFs.TomlConfig))

	return infoNamespace.createTable(L)
}

func getInfoNamespace(importLoopData *shared.ImportLoopData, L *lua.LState) lua.LValue {
	infoApi := importLoopData.InfoApi
	infoNamespace := newLuaNamespace()

	infoNamespace.AddFn("log", infoApi.Log)
	infoNamespace.AddFn("debug", infoApi.Debug)
	infoNamespace.AddFn("warn", infoApi.Warn)
	infoNamespace.AddFn("error", infoApi.Error)
	infoNamespace.AddFn("important", infoApi.Important)

	return infoNamespace.createTable(L)
}

func getShNamespace(L *lua.LState) lua.LValue {
	shellNamespace := newLuaNamespace()

	shellNamespace.AddFn("command", api.ShellCommand)

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
