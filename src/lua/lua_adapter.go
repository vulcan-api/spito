package lua

import (
	"github.com/Shopify/go-lua"
	"github.com/oleiade/reflections"
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

	default:
		t := reflect.ValueOf(field).Type()

		panic(
			"Struct, Array, Primitive or other thing\n" +
				"which you are trying to pass to lua Stack contains\n" +
				"type: " + t.Name() + " which is currently unsupported\n",
		)
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
