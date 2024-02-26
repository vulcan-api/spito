package userinfo

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
)

const RootEuid = 0

func ChangeToRoot() {
	err := syscall.Seteuid(RootEuid)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[error] run spito with sudo to continue")
		os.Exit(1)
	}
}

func GetRegularUser() *user.User {

	lognameCommand := exec.Command("logname")
	username, err := lognameCommand.Output()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	username = username[:len(username)-1] // remove trailing '\n' byte

	userObject, err := user.Lookup(string(username))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return userObject
}

func IsRoot() (bool, error) {
	currentUser, err := user.Current()
	return currentUser.Username == "root", err
}

func ChangeToUser() {

	userObject := GetRegularUser()
	uid, err := strconv.Atoi(userObject.Uid)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	err = syscall.Seteuid(uid)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
