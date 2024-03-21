package cmdApi

import "fmt"

type InfoApi struct{}

func (_ InfoApi) Log(args ...any) {
	_args := intoPrintArray("log", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Debug(args ...any) {
	_args := intoPrintArray("debug", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Error(args ...any) {
	_args := intoPrintArray("error", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Warn(args ...any) {
	_args := intoPrintArray("warn", args)
	fmt.Println(_args...)
}

func (_ InfoApi) Important(args ...any) {
	_args := intoPrintArray("important", args)
	fmt.Println(_args...)
}

func intoPrintArray(prefix string, args []any) []any {
	result := make([]any, len(args)+1)

	for i, e := range args {
		if i == 0 {
			result[i] = "[" + prefix + "]"
		}
		result[i] = fmt.Sprint(e)
	}

	return result
}
