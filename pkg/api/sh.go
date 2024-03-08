package api

import (
	"errors"
	"github.com/avorty/spito/pkg/userinfo"
	"os"
	"os/exec"
	"strconv"
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

	if strings.TrimSpace(command) == "" {
		return errors.New("command cannot be empty")
	}

	regularUser, err := userinfo.GetRegularUser()
	if err != nil {
		return err
	}

	uid, err := strconv.Atoi(regularUser.Uid)
	if err != nil {
		return err
	}

	err = syscall.Setreuid(uid, uid)
	if err != nil {
		return err
	}

	argv := strings.Split(command, " ")
	return syscall.Exec(argv[0], argv, os.Environ())
}
