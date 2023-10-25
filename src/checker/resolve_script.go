package checker

import (
	"fmt"
	"os"
	"strings"
)

func getScript(path string, ruleName string) (string, error) {
	scriptPath, err := getSpecificRulePath(path, ruleName)
	if err != nil {
		return "", err
	}

	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}
	return string(script), nil
}

func getSpecificRulePath(repoPath string, ruleName string) (string, error) {
	spitoRulesDataBytes, err := os.ReadFile(repoPath + "/SPITO_RULES")
	if err != nil {
		return "", err
	}

	spitoRulesData := strings.TrimSpace(string(spitoRulesDataBytes))
	lines := strings.Split(spitoRulesData, "\n")

	for _, line := range lines{
		if len(line) == 0 || strings.TrimSpace(line)[0] == '#' {
			continue
		}
		
		sides := strings.Split(line, "=")

		name := strings.TrimSpace(sides[0])
		path := strings.TrimSpace(sides[1])
		
		if ruleName == name {
			if path[0:2] == "./" {
				path = path[1:]
			} else if path[0] != '/' {
				path = "/" + path
			}
			path = repoPath + path
			
			return path, nil
		}
	}

	return "", fmt.Errorf("NOT FOUND rule called: " + ruleName)
}
