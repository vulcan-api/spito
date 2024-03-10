package api

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func ShellCommand(script string) (string, error) {
	cmd := exec.Command("sh", "-c", script)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

func splitArgs(command string) []string {
	command = strings.TrimSpace(command)
	var argv []string
	var currentArg string
	previousQuote := int32(0)

	isInQuotes := false

	for _, char := range command {
		isOpeningQuote := previousQuote == 0 && (char == '"' || char == '\'')
		isClosingQuote := (previousQuote == '"' && char == '"') || (previousQuote == '\'' && char == '\'')
		switch {
		case isOpeningQuote || isClosingQuote:
			isInQuotes = !isInQuotes
			if isOpeningQuote {
				previousQuote = char
			} else {
				previousQuote = 0
			}
			break
		case char == ' ' && !isInQuotes:
			argv = append(argv, currentArg)
			currentArg = ""
			break
		default:
			currentArg += string(char)
		}
	}

	if currentArg != "" {
		argv = append(argv, currentArg)
	}
	return argv
}

func Exec(command string) error {
	if strings.TrimSpace(command) == "" {
		return errors.New("command cannot be empty")
	}

	argv := splitArgs(command)
	return syscall.Exec(argv[0], argv, os.Environ())
}
