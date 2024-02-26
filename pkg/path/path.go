package path

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"github.com/avorty/spito/pkg/userinfo"
)

const (
	DirectoryPermissions = 0755
	FilePermissions      = 0644
)

var UserHomeDir = func() string {
	user := userinfo.GetRegularUser()

	// Exiting our program is only allowed because this function
	// executes only once, at the program start
	return user.HomeDir
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
	usr := userinfo.GetRegularUser()

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
