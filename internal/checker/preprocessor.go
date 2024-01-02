package checker

import (
	"github.com/avorty/spito/pkg/api"
	"regexp"
	"strings"
	"unicode"
)

const rulesetDirConstantName = "@ruleset"

func processScript(script string, ruleConf *RuleConf, rulesetPath string) string {
	newScript, decorators := getDecorators(script)

	newScript = strings.ReplaceAll(newScript, rulesetDirConstantName,  rulesetPath)

	for _, decorator := range decorators {
		if strings.ToLower(decorator) == "unsafe" {
			ruleConf.Unsafe = true
		}
	}

	return newScript
}

// Returns script without decorators and array of decorator values
func getDecorators(script string) (string, []string) {
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
