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
			return nil, err
		}
		options = append(options, processedOption)
		println(trimmedOptions)
	}
	fmt.Printf("%+v\n%s\n\n", options, trimmedOptions)

	return nil, nil
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

	// Possible edge cases:
	// name:type=val
	// name=val
	// name:type
	// name
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
	processedOption.DefaultValue, err = ParseDefaultValue(rawDefaultValue, processedOption.Type)
	if err != nil {
		return processedOption, rawOptions, err
	}

	resultScript := ""
	if commaPos < len(rawOptions) {
		resultScript = rawOptions[commaPos+1:]
	}
	return processedOption, resultScript, nil
}

func GetOptionType(rawType string) (option.Type, error) {
	var optionType option.Type
	processedType := strings.ToLower(rawType)
	switch processedType {
	case "int", "number":
		optionType = option.Int
		break
	case "uint":
		optionType = option.UInt
		break
	case "float":
		optionType = option.Float
		break
	case "bool", "boolean":
		optionType = option.Bool
		break
	case "string", "text":
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
	default:
		parsedValue = rawValue
		if rawValue[0] == '"' && rawValue[len(rawValue)-1] == '"' {
			partiallyParsedValue := strings.TrimSuffix(rawValue, "\"")
			parsedValue = strings.TrimPrefix(partiallyParsedValue, "\"")
		}
		break
	}
	return parsedValue, err
}
