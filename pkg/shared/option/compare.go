package option

import (
	"fmt"
	"strings"
)

func Compare(userInput []string, realOptions []Option) ([]Option, error) {
	for _, userOption := range userInput {
		splitPos := strings.Index(userOption, "=")
		name := userOption[:splitPos]
		value := userOption[splitPos+1:]
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
