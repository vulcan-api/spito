package checker

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"os"
)

const TmpDir = "/tmp/spito-rules"

func FetchRuleSet(ruleUrlPath string) (error, string, func()) {
	dir, removeDir, err := createTempDir()
	if err != nil {
		return err, "", nil
	}

	// TODO: complete this one
	_, err = git.PlainClone(dir, false, &git.CloneOptions{
		URL: ruleUrlPath,
	})
	if err != nil {
		return err, "", nil
	}

	return nil, dir, removeDir
}

// Returns: directory path, remove temp dir function, error
func createTempDir() (string, func(), error) {
	err := os.MkdirAll(TmpDir, 0700)
	if err != nil {
		return "", nil, err
	}
	dir, err := os.MkdirTemp(TmpDir, "rule")
	if err != nil {
		return "", nil, err
	}

	return dir,
		func() {
			err := os.RemoveAll(dir)
			if err != nil {
				fmt.Println("Failed to delete temp dir (path): ", dir)
			}
		},
		nil
}
