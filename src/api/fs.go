package api

import (
	"os"
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

func ReadFile(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(file), nil
}
func ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}
