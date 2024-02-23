package option

import (
	"fmt"
	"strconv"
	"strings"
)

type Type uint

const (
	Any Type = iota
	Int
	UInt
	Float
	String
	Bool
	Array // unhandled
	Struct
	Enum // unhandled
	Unknown
)

func (t Type) String() string {
	switch t {
	case Any:
		return "any"
	case Int:
		return "int"
	case UInt:
		return "uint"
	case Float:
		return "float"
	case String:
		return "string"
	case Bool:
		return "bool"
	case Array:
		return "array"
	case Struct:
		return "struct"
	case Enum:
		return "enum"
	default:
		return "unknown"
	}
}

type Option struct {
	Name           string
	DefaultValue   any
	Type           Type
	Optional       bool
	PossibleValues []string
	Options        []Option
}

func GetType(rawValue any) Type {
	_, ok := rawValue.(int)
	if ok {
		return Int
	}
	_, ok = rawValue.(uint)
	if ok {
		return UInt
	}
	_, ok = rawValue.(float64)
	if ok {
		return Float
	}
	_, ok = rawValue.(bool)
	if ok {
		return Bool
	}
	_, ok = rawValue.(string)
	if ok {
		return String
	}
	return Unknown
}

func GetValueAndType(rawValue string) (any, Type) {
	var parsedValue any
	parsedValue, err := strconv.Atoi(rawValue)
	if err == nil {
		return parsedValue, Int
	}
	parsedValue, err = strconv.ParseUint(rawValue, 10, 0)
	if err == nil {
		return parsedValue, UInt
	}
	parsedValue, err = strconv.ParseFloat(rawValue, 10)
	if err == nil {
		return parsedValue, Float
	}
	parsedValue, err = strconv.ParseBool(rawValue)
	if err == nil {
		return parsedValue, Bool
	}
	rawValue = fmt.Sprint(rawValue)
	parsedValue = rawValue
	if len(rawValue) > 0 && rawValue[0] == '"' && rawValue[len(rawValue)-1] == '"' {
		partiallyParsedValue := strings.TrimSuffix(rawValue, "\"")
		parsedValue = strings.TrimPrefix(partiallyParsedValue, "\"")
	}
	return parsedValue, String
}
