package shared

import (
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	DirectoryPermissions = 0755
	FilePermissions      = 0644
)

var UserHomeDir = func() string {
	dir, err := os.UserHomeDir()

	// Exiting our program is only allowed because this function
	// executes only once, at the program start
	if err != nil {
		fmt.Println("Failed to read UserHomeDir\n", err.Error())
		os.Exit(1)
	}
	return dir
}()

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateIfNotExists(path, defaultContent string) error {
	pathExists, err := PathExists(path)
	if err != nil {
		return err
	}

	if !pathExists {
		if err := os.MkdirAll(filepath.Dir(path), DirectoryPermissions); err != nil {
			return err
		}

		file, err := os.Create(path)
		if err != nil {
			return err
		}

		if _, err := file.WriteString(defaultContent); err != nil {
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func ExpandTilde(path *string) error {
	usr, err := user.LookupId(strconv.Itoa(syscall.Geteuid()))
	if err != nil {
		return err
	}

	if *path == "~" {
		*path = usr.HomeDir
	}
	if strings.HasPrefix(*path, "~/") {
		*path = strings.Replace(*path, "~", usr.HomeDir, 1)
	}
	return nil
}

func GetEnvWithDefaultValue(environmentVariable string, defaultValue string) string {
	if val := os.Getenv(environmentVariable); val != "" {
		return val
	}
	return defaultValue
}

func RandomLetters(length int) string {
	letters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}

	return string(result)
}
