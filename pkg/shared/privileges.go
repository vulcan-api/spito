package shared

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

const RootEuid = 0
const sudoUsernameVariable = "SUDO_USER"

func ChangeToRoot() {
	err := syscall.Seteuid(RootEuid)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "[error] run spito with root privileges to continue")
		os.Exit(1)
	}
}

func GetRegularUser() *user.User {
	userObject, err := user.Lookup(os.Getenv(sudoUsernameVariable))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return userObject
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
