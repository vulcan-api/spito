package api

import (
	"os"
	"strings"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func FileExists(path string, isDirectory bool) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	if isDirectory && info.IsDir() {
		return true
	}
	if !isDirectory && !info.IsDir() {
		return true
	}

	return false
}
func FileContains(path string, content string) (bool, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	str := string(file)
	contains := strings.Contains(str, content)

	return contains, nil
}
