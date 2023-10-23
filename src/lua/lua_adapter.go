package lua

import (
	"fmt"
	"github.com/Shopify/go-lua"
	"github.com/oleiade/reflections"
	"go/types"
	"reflect"
	"strings"
)

func AddStructToStack(L *lua.State, obj interface{}) error {
	fields, err := reflections.Fields(obj)
	if err != nil {
		return err
	}

	// Create key-value lua Table
	L.CreateTable(0, len(fields))

	for _, fieldName := range fields {
		// Create new field
		L.PushString(strings.ToLower(fieldName))

		field, err := reflections.GetField(obj, fieldName)
		kind, err := reflections.GetFieldKind(obj, fieldName)

		if err != nil {
			return err
		}
		if err := addToStackBasedOnType(L, field, kind); err != nil {
			return err
		}

		// Set key and value to lua Table
		L.RawSet(-3)
	}

	methods := reflect.ValueOf(obj)
	for i := 0; i < methods.NumMethod(); i++ {
		method := methods.Method(i)
		_ = method
		// TODO: implement methods
	}

	return nil
}

func AddArrayToStack(L *lua.State, array interface{}) error {
	fields := reflect.ValueOf(array)

	// Create number-indexed lua Table (like array)
	L.CreateTable(fields.Len(), 0)

	for i := 0; i < fields.Len(); i++ {
		field := fields.Index(i)
		if err := addToStackBasedOnType(L, field, field.Kind()); err != nil {
			return err
		}
		// Insert value to lua Table
		L.RawSet(-2)
	}

	return nil
}

func AddToStackPrimitiveType(L *lua.State, field interface{}) {
	switch field.(type) {
	case string:
		L.PushString(field.(string))
		break

	case int:
		L.PushInteger(field.(int))
		break

	case uint:
		L.PushInteger(int(field.(uint)))
		break

	case float64:
		L.PushNumber(field.(float64))
		break

	case float32:
		L.PushNumber(float64(field.(float32)))
		break

	case bool:
		L.PushBoolean(field.(bool))
		break

	case error:
		L.PushString(field.(error).Error())
		break

	case types.Nil:
		L.PushNil()
		break

	default:
		fieldVal := reflect.ValueOf(field)
		var err error
		
		if !fieldVal.IsValid() {
			err = fmt.Errorf(
				"Struct, Array, Primitive or other thing\n" +
					"which you are trying to pass to lua Stack is not valid (probably reflect zero Value)\n",
			)
		} else {
			err = fmt.Errorf(
				"Struct, Array, Primitive or other thing\n" +
					"which you are trying to pass to lua Stack contains\n" +
					"type: " + fieldVal.Type().Name() + " which is currently unsupported\n",
			)
		}
		
		fmt.Println(err.Error())
		L.PushNil()
	}
}

func addToStackBasedOnType(L *lua.State, field interface{}, kind reflect.Kind) error {
	if kind == reflect.Array {
		if err := AddArrayToStack(L, field); err != nil {
			return err
		}
		return nil
	}

	if kind == reflect.Struct {
		err := AddStructToStack(L, field)
		if err != nil {
			return err
		}
		return nil
	}

	// I haven't found any way to check whether a field kind is Primitive
	// So it potentially can cause unexpected errors
	AddToStackPrimitiveType(L, field)
	return nil
}

func pushMethodToStack(L *lua.State, obj *reflect.Value, method reflect.Method, done func()) {
	obj.Type()
	methodType := method.Type
	L.PushGoFunction(func(state *lua.State) int {
		var args []interface{}
		for i := 1; i < methodType.NumIn(); i++ {
			args = append(args, state.ToValue(i))
		}

		//args = append([]reflect.Value{toStructPtr(obj)}, args...)
		//results := method.Func.Call(args)
		results, err := invokeMethod(obj, method.Name, args...)
		if err != nil {
			println(err.Error())
			return 0
		}

		for _, result := range results {
			err := addToStackBasedOnType(state, result.Interface(), result.Kind())
			if err != nil {
				return 0
			}
		}

		done()
		return len(results)
	})
}

func toStructPtr(obj reflect.Value) reflect.Value {
	vp := reflect.New(obj.Type())
	vp.Elem().Set(obj)
	//vp = vp.Convert(obj.Type())
	//*obj = vp
	println(vp.Type().String())
	return vp
}

func invokeMethod(obj *reflect.Value, name string, args ...interface{}) ([]reflect.Value, error) {
	objPtr := toStructPtr(*obj)
	isObjPtrInUse := false
	var method reflect.Value
	
	if m := objPtr.MethodByName(name); m.IsValid() {
		isObjPtrInUse = true
		method = m
	} else if m := obj.MethodByName(name); m.IsValid() {
		method = m
	} else {
		return []reflect.Value{}, fmt.Errorf("Method not found")
	}

	methodType := method.Type()
	numIn := methodType.NumIn()
	if numIn > len(args) {
		return []reflect.Value{}, fmt.Errorf("Method %s must have minimum %d params. Have %d", name, numIn, len(args))
	}
	if numIn != len(args) && !methodType.IsVariadic() {
		return []reflect.Value{}, fmt.Errorf("Method %s must have %d params. Have %d", name, numIn, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < len(args); i++ {
		var inType reflect.Type
		if methodType.IsVariadic() && i >= numIn-1 {
			inType = methodType.In(numIn - 1).Elem()
		} else {
			inType = methodType.In(i)
		}
		argValue := reflect.ValueOf(args[i])
		if !argValue.IsValid() {
			return []reflect.Value{}, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argValue.String())
		}
		argType := argValue.Type()
		if argType.ConvertibleTo(inType) {
			in[i] = argValue.Convert(inType)
		} else {
			return []reflect.Value{}, fmt.Errorf("Method %s. Param[%d] must be %s. Have %s", name, i, inType, argType)
		}
	}
	res := method.Call(in)

	if isObjPtrInUse {
		*obj = objPtr.Elem()
	}
	
	return res, nil
}
