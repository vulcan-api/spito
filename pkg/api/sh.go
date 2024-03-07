package api

import (
	"os"
	"os/exec"
	"strings"
	"path/filepath"
	"fmt"
	"github.com/avorty/spito/pkg/userinfo"
	"syscall"
	"errors"
	"strconv"
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

	const executablesDirectory = "/usr/bin"
	argv := strings.Split(command, " ")
	argv[0] = filepath.Join(executablesDirectory, command)

	err = syscall.Exec(argv[0], argv, os.Environ())
	fmt.Println(err.Error())
	return err
}
