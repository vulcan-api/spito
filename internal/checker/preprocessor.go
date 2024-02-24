package checker

import (
	"fmt"
	"github.com/avorty/spito/pkg/api"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/shared/option"
	"regexp"
	"strings"
	"unicode"
)

type DecoratorType uint

const (
	UnsafeDecorator = iota
	DescriptionDecorator
	OptionsDecorator
	EnvironmentDecorator
	SudoDecorator
	UnknownDecorator
)

type RawDecorator struct {
	Type    DecoratorType
	Content string
}

func processScript(script string, ruleConf *shared.RuleConfigLayout) (string, error) {
	newScript, decorators, err := GetDecorators(script)
	if err != nil {
		return newScript, err
	}

	for _, decorator := range decorators {
		switch decorator.Type {
		case UnsafeDecorator:
			ruleConf.Unsafe = true
			break
		case OptionsDecorator:
			ruleConf.Options, err = option.AppendOptions(ruleConf.Options, decorator.Content)
			if err != nil {
				return newScript, err
			}
			break
		case EnvironmentDecorator:
			ruleConf.Environment = true
			break
		case SudoDecorator:
			ruleConf.Sudo = true
			break
		default:
			break
		}
	}

	return newScript, nil
}

// GetDecorators Returns script without decorators and array of decorator values
func GetDecorators(script string) (string, []RawDecorator, error) {
	var fileScopeDecorators []RawDecorator

	fileScopeRegex := regexp.MustCompile(`#!\[[^]]+]`)
	decoratorMatches := fileScopeRegex.FindAllString(script, -1)

	for _, decorator := range decoratorMatches {
		script = strings.Replace(script, decorator, "", 1)

		processedDecorator := api.RemoveComments(decorator, "--", "--[[", "]]")
		processedDecorator = removeWhitespaces(processedDecorator)

		processedDecorator = strings.TrimPrefix(processedDecorator, "#![")
		processedDecorator = strings.TrimSuffix(processedDecorator, "]")

		var finalDecorator RawDecorator
		bracketIndex := strings.Index(processedDecorator, "(")
		rawDecoratorName := processedDecorator
		if bracketIndex != -1 {
			rawDecoratorName = processedDecorator[0:bracketIndex]
		}

		var err error
		finalDecorator.Type, err = GetDecoratorType(rawDecoratorName)
		if err != nil {
			return script, fileScopeDecorators, err
		}

		betweenParenthesesRegex := regexp.MustCompile(`\(.*\)`)
		decoratorContent := betweenParenthesesRegex.FindString(processedDecorator)
		if len(decoratorContent) > 2 {
			decoratorContent = strings.TrimPrefix(decoratorContent, "(")
			decoratorContent = strings.TrimSuffix(decoratorContent, ")")
		}
		finalDecorator.Content = decoratorContent

		fileScopeDecorators = append(fileScopeDecorators, finalDecorator)
	}

	return script, fileScopeDecorators, nil
}

func GetDecoratorType(name string) (DecoratorType, error) {
	var decoratorType DecoratorType

	processedName := strings.ToLower(name)
	switch processedName {
	case "unsafe":
		decoratorType = UnsafeDecorator
		break
	case "description":
		decoratorType = DescriptionDecorator
		break
	case "options":
		decoratorType = OptionsDecorator
	case "environment":
		decoratorType = EnvironmentDecorator
		break
	case "sudo":
		decoratorType = SudoDecorator
	default:
		return UnknownDecorator, fmt.Errorf("unknown decorator: %s", name)
	}
	return decoratorType, nil
}

func removeQuotes(text *string) {
	*text = strings.TrimPrefix(*text, "\"")
	*text = strings.TrimSuffix(*text, "\"")
}

func GetDecoratorArguments(decoratorCode string) ([]string, map[string]string, error) {
	// TODO: escape it first
	arguments := strings.Split(decoratorCode, ",")

	var positionalArguments []string
	namedArguments := make(map[string]string)

	for argumentIndex, argument := range arguments {
		argumentTokens := strings.Split(argument, "=")
		if argTokenLen := len(argumentTokens); argTokenLen > 2 {
			return nil, nil, fmt.Errorf("syntax error in argument number %d", argumentIndex)
		} else if argTokenLen == 1 {
			removeQuotes(&argumentTokens[0])

			positionalArguments = append(positionalArguments, argumentTokens[0])
			continue
		}
		removeQuotes(&argumentTokens[1])
		namedArguments[argumentTokens[0]] = argumentTokens[1]
	}
	return positionalArguments, namedArguments, nil
}

func removeWhitespaces(decorator string) string {
	var result strings.Builder
	isString := false

	for i := 0; i < len(decorator); i++ {
		char := decorator[i]

		if char == '"' {
			isString = !isString
		}
		if !unicode.IsSpace(rune(char)) || isString {
			result.WriteByte(char)
		}
	}

	return result.String()
}
