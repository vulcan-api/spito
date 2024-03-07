package api

import (
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

func Exec(command string) error {
	return syscall.Exec(command, strings.Split(command, " "), os.Environ())
}
