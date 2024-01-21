package checker

import (
	"errors"
	"fmt"
	"github.com/avorty/spito/pkg/api"
	"regexp"
	"strings"
	"unicode"
)

func processScript(script string, ruleConf *RuleConf) string {
	newScript, decorators := GetDecorators(script)

	for _, decorator := range decorators {
		if strings.ToLower(decorator) == "unsafe" {
			ruleConf.Unsafe = true
		}
	}

	return newScript
}

// Returns script without decorators and array of decorator values
func GetDecorators(script string) (string, []string) {
	var fileScopeDecorators []string

	fileScopeRegex := regexp.MustCompile(`#!\[[^]]+]`)
	decoratorMatches := fileScopeRegex.FindAllString(script, -1)

	for _, decorator := range decoratorMatches {
		script = strings.Replace(script, decorator, "", 1)

		decorator = api.RemoveComments(decorator, "--", "--[[", "]]")
		decorator = removeWhitespaces(decorator)

		decorator = strings.TrimPrefix(decorator, "#![")
		decorator = strings.TrimSuffix(decorator, "]")

		fileScopeDecorators = append(fileScopeDecorators, decorator)
	}

	return script, fileScopeDecorators
}

func GetDecoratorArguments(decoratorCode string) ([]string, map[string]string, error) {
	betweenParenthesesRegex := regexp.MustCompile(`\(.*\)`)
	argumentCode := betweenParenthesesRegex.FindString(decoratorCode)
	argumentCode = strings.TrimPrefix(argumentCode, "(")
	argumentCode = strings.TrimSuffix(argumentCode, ")")

	arguments := strings.Split(argumentCode, ",")

	var positionalArguments []string
	var namedArguments map[string]string

	for argumentNumber, argument := range arguments {
		argumentTokens := strings.Split(argument, "=")
		if len(argumentTokens) > 2 {
			return nil, nil, errors.New(fmt.Sprintf("syntax error in argument number %d", argumentNumber))
		}

		if len(argumentTokens) == 1 {
			argumentTokens[0] = strings.TrimPrefix(argumentTokens[0], "\"")
			argumentTokens[0] = strings.TrimSuffix(argumentTokens[0], "\"")

			positionalArguments = append(positionalArguments, argumentTokens[0])
			continue
		}
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
