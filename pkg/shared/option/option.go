package option

import (
	"fmt"
	"strconv"
	"strings"
)

type Option struct {
	Name           string
	DefaultValue   any
	Type           Type
	Optional       bool
	PossibleValues []string
	Options        []Option
}

func (o *Option) SetValue(rawValue string) error {
	parsedValue, err := Parse(rawValue, o.Type)
	if err != nil && o.Type == Enum {
		foundMatch := false
		for _, possibleValue := range o.PossibleValues {
			if possibleValue == rawValue {
				foundMatch = true
				break
			}
		}
		if !foundMatch {
			return fmt.Errorf("passed value doesn't suite '%s' enum's possible values (%+v): '%s'", o.Name, o.PossibleValues, rawValue)
		}
	} else if err != nil {
		return err
	}

	o.DefaultValue = parsedValue
	return nil
}

type Type uint

const (
	Any Type = iota
	Int
	UInt
	Float
	String
	Bool
	List
	Struct
	Enum
	Unknown
)

var ParsableTypes = []Type{Int, UInt, Float, Bool}

func FromString(rawType string) Type {
	switch rawType {
	case "any":
		return Any
	case "int":
		return Int
	case "uint":
		return UInt
	case "float":
		return Float
	case "string":
		return String
	case "bool":
		return Bool
	case "list":
		return List
	default:
		return Unknown
	}
}

func (t Type) ToString() string {
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
	case List:
		return "list"
	case Struct:
		return "struct"
	case Enum:
		return "enum"
	default:
		return "unknown"
	}
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
	var err error
	for _, parsableType := range ParsableTypes {
		parsedValue, err = Parse(rawValue, parsableType)
		if err == nil {
			return parsedValue, Int
		}
	}
	parsedValue = rawValue
	if len(rawValue) > 0 && rawValue[0] == '"' && rawValue[len(rawValue)-1] == '"' {
		partiallyParsedValue := strings.TrimSuffix(rawValue, "\"")
		parsedValue = strings.TrimPrefix(partiallyParsedValue, "\"")
	}
	return parsedValue, String
}

func Parse(rawValue string, valueType Type) (any, error) {
	var parsedValue any
	var err error
	switch valueType {
	case Int:
		parsedValue, err = strconv.Atoi(rawValue)
		break
	case UInt:
		parsedValue, err = strconv.ParseUint(rawValue, 10, 0)
		break
	case Float:
		parsedValue, err = strconv.ParseFloat(rawValue, 10)
		break
	case Bool:
		parsedValue, err = strconv.ParseBool(rawValue)
		break
	case String:
		parsedValue = rawValue
		break
	default:
		return nil, fmt.Errorf("value '%s' cannot be parsed to option of type '%s': unsupported type", rawValue, valueType.ToString())
	}
	if err != nil {
		return nil, err
	}

	return parsedValue, nil
}
