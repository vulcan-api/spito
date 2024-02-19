package checker

import (
	"fmt"
	"github.com/avorty/spito/pkg/shared/option"
	"strconv"
	"strings"
)

func AppendOptions(originalOptions []option.Option, rawOptions string) ([]option.Option, error) {
	newOptions, err := ParseOptions(rawOptions)
	if err != nil {
		return originalOptions, err
	}
	mergedOptions := append(originalOptions, newOptions...)
	return mergedOptions, nil
}

func ParseOptions(rawOptions string) ([]option.Option, error) {
	trimmedOptions := rawOptions[1 : len(rawOptions)-1]
	var options []option.Option
	var err error
	for len(trimmedOptions) > 0 {
		var processedOption option.Option
		processedOption, trimmedOptions, err = GetOption(trimmedOptions)
		if err != nil {
			return options, err
		}
		options = append(options, processedOption)
	}

	return options, nil
}

func GetOption(rawOptions string) (option.Option, string, error) {
	processedOption := option.Option{}

	commaPos := strings.Index(rawOptions, ",")
	colonPos := strings.Index(rawOptions, ":")
	equalPos := strings.Index(rawOptions, "=")

	if commaPos == -1 {
		commaPos = len(rawOptions)
	}

	foundColon := colonPos != -1 && colonPos < commaPos
	foundEqual := equalPos != -1 && equalPos < commaPos

	var err error

	// Possible HANDLED edge cases:
	// name:type=val
	// name=val
	// name:type
	// name

	// Possible UNHANDLED cases:
	//
	// name:array=[false, "one", 2]
	// name=[false, "one", 2,]
	//
	// name:struct={name:type=val}
	// name={name:type=val}
	//
	// name?:type=val
	// name?:type
	// name?=val
	// name?
	//
	// ""
	processedOption.Type = option.Any
	rawDefaultValue := ""

	if foundColon {
		processedOption.Name = rawOptions[0:colonPos]
		typeSlice := rawOptions[colonPos+1 : commaPos]
		if foundEqual {
			typeSlice = rawOptions[colonPos+1 : equalPos]
			rawDefaultValue = rawOptions[equalPos+1 : commaPos]
		}
		processedOption.Type, err = GetOptionType(typeSlice)
		if err != nil {
			return processedOption, rawOptions, err
		}
	} else if foundEqual {
		processedOption.Name = rawOptions[0:equalPos]
		rawDefaultValue = rawOptions[equalPos+1 : commaPos]
	} else {
		processedOption.Name = rawOptions[0:commaPos]
	}

	obtainedDefaultValueType := processedOption.Type
	if processedOption.Type == option.Any {
		obtainedDefaultValueType = option.GetType(rawDefaultValue)
		if processedOption.Type == option.Unknown {
			return processedOption, rawOptions, fmt.Errorf("cannot convert option '%s' to any recognized type", processedOption.Name)
		}
	}
	processedOption.DefaultValue, err = ParseDefaultValue(rawDefaultValue, obtainedDefaultValueType)
	if err != nil {
		return processedOption, rawOptions, err
	}

	name := processedOption.Name
	lastChar := name[len(name)-1]

	if lastChar == '?' {
		processedOption.Optional = true
	}

	resultScript := ""
	if commaPos < len(rawOptions) {
		resultScript = rawOptions[commaPos+1:]
	}

	if !processedOption.Optional && processedOption.DefaultValue == nil {
		return processedOption, resultScript, fmt.Errorf("option '%s' must be optional or provide default value", processedOption.Name)
	}

	return processedOption, resultScript, nil
}

func GetOptionType(rawType string) (option.Type, error) {
	var optionType option.Type
	processedType := strings.ToLower(rawType)
	switch processedType {
	case "int":
		optionType = option.Int
		break
	case "uint":
		optionType = option.UInt
		break
	case "float":
		optionType = option.Float
		break
	case "bool":
		optionType = option.Bool
		break
	case "string":
		optionType = option.String
		break
	// TODO: requires more effort
	//case "array", "list":
	//	optionType = option.Array
	//	break
	//case "struct", "object":
	//	optionType = option.Struct
	//case "enum":
	//	optionType = option.Struct
	default:
		return option.Unknown, fmt.Errorf("unknown option type in Options decorator: '%s'", rawType)
	}
	return optionType, nil
}

func ParseDefaultValue(rawValue string, valueType option.Type) (any, error) {
	var parsedValue any
	var err error
	switch valueType {
	case option.Int:
		parsedValue, err = strconv.Atoi(rawValue)
		break
	case option.UInt:
		parsedValue, err = strconv.ParseUint(rawValue, 10, 0)
		break
	case option.Float:
		parsedValue, err = strconv.ParseFloat(rawValue, 10)
		break
	case option.Bool:
		parsedValue, err = strconv.ParseBool(rawValue)
		break
	case option.String:
		parsedValue = rawValue
		if len(rawValue) > 0 && rawValue[0] == '"' && rawValue[len(rawValue)-1] == '"' {
			partiallyParsedValue := strings.TrimSuffix(rawValue, "\"")
			parsedValue = strings.TrimPrefix(partiallyParsedValue, "\"")
		}
		break
	default:
		err = fmt.Errorf("unsupported option type '%s' of value '%s'", valueType, rawValue)
		break
	}
	return parsedValue, err
}
