package shared

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
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

func DoesPathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CreateIfNotExist(path, defaultContent string) error {
	pathExist, err := DoesPathExist(path)
	if err != nil {
		return err
	}

	if !pathExist {
		if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
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
	usr, err := user.Current()
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
