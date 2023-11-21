package cmdApi

import "fmt"

type InfoApi struct{}

func (_ InfoApi) Log(args ...string) {
	_args := intoPrintArray("[log]", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Debug(args ...string) {
	_args := intoPrintArray("[debug]", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Error(args ...string) {
	_args := intoPrintArray("[error]", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Warn(args ...string) {
	_args := intoPrintArray("[warn]", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Important(args ...string) {
	_args := intoPrintArray("[important]", args)
	fmt.Println(_args...)
}

func intoPrintArray(prefix string, args []string) []any {
	result := make([]any, len(args)+1)
	result[0] = prefix

	for i, e := range args {
		result[i+1] = e
	}

	return result
}
