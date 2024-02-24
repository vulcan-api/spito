package option

import (
	"fmt"
	"strings"
)

func AppendOptions(originalOptions []Option, rawOptions string) ([]Option, error) {
	newOptions, err := ParseOptions(rawOptions)
	if err != nil {
		return originalOptions, err
	}
	mergedOptions := append(originalOptions, newOptions...)
	return mergedOptions, nil
}

func ParseOptions(rawOptions string) ([]Option, error) {
	trimmedOptions := rawOptions
	if len(rawOptions) > 2 {
		trimmedOptions = rawOptions[1 : len(rawOptions)-1]
	}
	options := make([]Option, 0)
	var err error
	for len(trimmedOptions) > 0 {
		var processedOption Option
		processedOption, trimmedOptions, err = GetOption(trimmedOptions)
		if err != nil {
			return options, err
		}
		options = append(options, processedOption)
	}

	return options, nil
}

func GetOption(rawOptions string) (Option, string, error) {
	processedOption := Option{}

	commaPos, err := GetIndexOutside(rawOptions, "{", "}", ",")
	if err != nil {
		return processedOption, rawOptions, err
	}
	colonPos, err := GetIndexOutside(rawOptions, "{", "}", ":")
	if err != nil {
		return processedOption, rawOptions, err
	}
	equalPos, err := GetIndexOutside(rawOptions, "{", "}", "=")
	if err != nil {
		return processedOption, rawOptions, err
	}
	if commaPos == -1 {
		commaPos = len(rawOptions)
	}

	foundColon := colonPos != -1 && colonPos < commaPos
	foundEqual := equalPos != -1 && equalPos < commaPos

	// Possible cases:
	// name:type=val
	// name?:type
	// name=val
	// name?
	//
	// name={name:type=val}
	// enum:{FOOD;CAT;4}=FOOD

	// Illegal cases:
	// name:type
	// name

	processedOption.Type = Any
	rawDefaultValue := ""

	if foundColon {
		processedOption.Name = rawOptions[0:colonPos]
		typeSlice := rawOptions[colonPos+1 : commaPos]
		if foundEqual {
			typeSlice = rawOptions[colonPos+1 : equalPos]
			rawDefaultValue = rawOptions[equalPos+1 : commaPos]
		}
		processedOption.Type, processedOption.PossibleValues, err = GetOptionType(typeSlice)
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
	if processedOption.Type == Any {
		if IsArray(rawDefaultValue) {
			obtainedDefaultValueType = List
		} else if IsStruct(rawDefaultValue) {
			obtainedDefaultValueType = Struct
		}
	}

	switch obtainedDefaultValueType {
	case List:
		processedOption.DefaultValue = UnwrapArray(rawDefaultValue)
		break
	case Struct:
		processedOption.Options, err = ParseOptions(rawDefaultValue)
		if err != nil {
			return processedOption, rawOptions, fmt.Errorf("cannot convert option '%s' to a struct: %s", processedOption.Name, err.Error())
		}
		processedOption.Type = Struct
	default:
		if rawDefaultValue == "" {
			processedOption.DefaultValue = nil
		} else {
			processedOption.DefaultValue, _ = GetValueAndType(rawDefaultValue)
		}
		if err != nil {
			return processedOption, rawOptions, err
		}
	}

	name := processedOption.Name
	lastChar := name[len(name)-1]

	if lastChar == '?' {
		processedOption.Optional = true
		processedOption.Name = processedOption.Name[:len(processedOption.Name)-1]
	}

	resultScript := ""
	if commaPos < len(rawOptions) {
		resultScript = rawOptions[commaPos+1:]
	}

	if processedOption.Type != Struct && !processedOption.Optional && processedOption.DefaultValue == nil {
		return processedOption, resultScript, fmt.Errorf("option '%s' must be optional or provide default value", processedOption.Name)
	}

	return processedOption, resultScript, nil
}

func GetOptionType(rawType string) (Type, []string, error) {
	var optionType Type
	var possibleValues []string

	optionType = FromString(strings.ToLower(rawType))
	if optionType == Unknown {
		if isStruct := IsStruct(rawType); isStruct {
			optionType = Enum
			possibleValues = UnwrapArray(rawType)
		} else {
			return Unknown, nil, fmt.Errorf("unknown option type in options: '%s'", rawType)
		}
	}
	return optionType, possibleValues, nil
}

func IsStruct(rawValue string) bool {
	if rawValueLen := len(rawValue); rawValueLen > 2 {
		if rawValue[0] == '{' && rawValue[rawValueLen-1] == '}' {
			return true
		}
	}
	return false
}

func IsArray(rawValue string) bool {
	return IsStruct(rawValue) && !strings.ContainsAny(rawValue, ":=,")
}

func UnwrapArray(rawValue string) []string {
	if len(rawValue) > 2 {
		return strings.Split(rawValue[1:len(rawValue)-1], ";")
	}
	return nil
}

func GetIndexOutside(wholeString, toEscapeStart, toEscapeEnd, toFind string) (int, error) {
	indentNum := 1

	findPos := strings.Index(wholeString, toFind)

	if findPos == -1 {
		return -1, nil
	}

	startPos := strings.Index(wholeString, toEscapeStart)
	endPos := strings.Index(wholeString, toEscapeEnd)

	if startPos == -1 || findPos < startPos || (endPos < startPos && findPos < endPos) {
		return findPos, nil
	}

	i := endPos
	if startPos < endPos {
		i = startPos
	}

	for indentNum > 0 || i >= len(wholeString) {
		if endPos == -1 {
			return -1, fmt.Errorf("unclosed structure (missing '}') in: >>>\n'%s'\n<<<", wholeString)
		}
		if startPos == endPos {
			return -1, fmt.Errorf("opening string (%s) is equal to the closing one (%s)", toEscapeStart, toEscapeEnd)
		}

		newStartPos := strings.Index(wholeString[i+1:], toEscapeStart)
		newEndPos := strings.Index(wholeString[i+1:], toEscapeEnd)

		foundStart := false
		if newStartPos != -1 {
			foundStart = true
		}

		foundEnd := false
		if newEndPos != -1 {
			foundEnd = true
		}

		startIsCloser := false
		if (foundStart && foundEnd && newStartPos < newEndPos) || foundStart && !foundEnd {
			startIsCloser = true
		}

		endIsCloser := false
		if (foundStart && foundEnd && newEndPos < newStartPos) || !foundStart && foundEnd {
			endIsCloser = true
		}

		if newEndPos == -1 {
			break
		}

		if startIsCloser {
			indentNum++
			i += newStartPos + 1
		} else if endIsCloser {
			i += newEndPos + 1
			indentNum--
		} else {
			break
		}
	}
	obtained := strings.Index(wholeString[i:], toFind)
	if obtained == -1 {
		return -1, nil
	}
	return i + obtained, nil
}
