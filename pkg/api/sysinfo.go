package api

import (
	"math/rand"
	"os/exec"
	"strings"
	"time"
	"os"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/zcalusic/sysinfo"
)

/* TYPES */

type Distro struct {
	Name    string
	Version string
}

type InitSystem string

func (is InitSystem) String() string {
	switch is {
	case SYSTEMD:
		return "systemd"
	case RUNIT:
		return "runit"
	}
	return ""
}

func isOpenRC() bool {
	const rcCommand = "rc-status"
	cmd := exec.Command(rcCommand)
	_, err := cmd.Output()

	if err != nil {
		return false
	}

	return true
}

/* CONSTANTS */

const (
	SYSTEMD InitSystem = "systemd"
	RUNIT   InitSystem = "runit"
	OPENRC  InitSystem = "openrc"
	/*
		lack of sysv, s6 and 66
	*/
	UNKNOWN InitSystem = ""
)

/* API FUNCTIONS */

func Sleep(milliseconds int) {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
}

func GetDistro() Distro {
	var systemInfo sysinfo.SysInfo
	systemInfo.GetSysInfo()

	return Distro{systemInfo.OS.Name, systemInfo.OS.Release}
}

func GetInitSystem() (InitSystem, error) {
	initSystemProcess, err := process.NewProcess(1)
	if err != nil {
		return "", err
	}

	processName, err := initSystemProcess.Name()
	if err != nil {
		return "", err
	}

	if strings.Contains(processName, SYSTEMD.String()) {
		return SYSTEMD, nil
	}
	if strings.Contains(processName, RUNIT.String()) {
		return RUNIT, nil
	}
	if strings.Contains(processName, "init") && isOpenRC() {
		return OPENRC, nil
	}

	return UNKNOWN, nil
}

func GetRandomLetters(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func GetEnv(variableName string) string {
	return os.Getenv(variableName)
}
