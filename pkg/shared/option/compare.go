package option

import (
	"fmt"
	"strings"
)

// Compare checks if raw array of strings e.g. "name=Linus,lastname=Torvalds" are applicable
// into rule's set of options e.g. {name?:string,lastname:int=0}
// Returns user modified array of options
func Compare(userInput []string, realOptions []Option) ([]Option, error) {
	for _, userOption := range userInput {
		name, value, properlyPassed := strings.Cut(userOption, "=")
		if properlyPassed != true {
			return realOptions, fmt.Errorf("passed option without value (missing '='): '%s'", value)
		}
		foundOption := false
		for i, realOption := range realOptions {
			if realOption.Name == name {
				err := realOptions[i].SetValue(value)
				foundOption = true
				if err != nil {
					return realOptions, err
				}
			}
		}
		if !foundOption {
			return realOptions, fmt.Errorf("no such option specified: '%s'", name)
		}
	}
	return realOptions, nil
}
