package guiApi

import "fmt"

type InfoApi struct{}

// Log TODO: use dbus
func (_ InfoApi) Log(args ...string) {
	fmt.Println(args)
}

func (_ InfoApi) Debug(args ...string) {
	fmt.Println(args)
}

func (_ InfoApi) Error(args ...string) {
	fmt.Println(args)
}

func (_ InfoApi) Warn(args ...string) {
	fmt.Println(args)
}

func (_ InfoApi) Important(args ...string) {
	fmt.Println(args)
}
