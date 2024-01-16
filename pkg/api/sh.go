package api

import "os/exec"

func ShellCommand(script string) (string, error) {
	cmd := exec.Command("sh", "-c", script)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
