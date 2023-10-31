package checker

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

func getRuleSetsDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return dir + "/.local/state/spito-rules/rule-sets", nil
}

func initRequiredTmpDirs() error {
	dir, err := getRuleSetsDir()
	if err != nil {
		return err
	}
	err = os.MkdirAll(dir, 0700)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	return err
}

func getDefaultRepoPrefix() string {
	// TODO: implement logic for getting default repo prefix
	return "github.com"
}

type RuleSetLocation struct {
	simpleUrl string
}

// e.g. from: https://github.com/Nasz-Elektryk/spito-ruleset.git to Nasz-Elektryk/spito-ruleset
func (r *RuleSetLocation) new(identifier string) {
	// check if simpleUrl is url:
	if !strings.Contains(identifier, ".") {
		r.simpleUrl = getDefaultRepoPrefix() + "/" + identifier
		return
	}

	simpleUrl := identifier
	simpleUrl = strings.ReplaceAll(simpleUrl, "https://", "")
	simpleUrl = strings.ReplaceAll(simpleUrl, "http://", "")
	simpleUrl = strings.ReplaceAll(simpleUrl, "www.", "")
	urlLen := len(simpleUrl)

	if simpleUrl[urlLen-1] == '/' {
		simpleUrl = simpleUrl[:urlLen-1]
	}
	// I still wonder whether it is good idea:
	if simpleUrl[urlLen-5:] == ".git" {
		simpleUrl = simpleUrl[urlLen-5:]
	}

	r.simpleUrl = simpleUrl
}

func (r *RuleSetLocation) createDir() error {
	println(r.getRuleSetPath())
	err := os.MkdirAll(r.getRuleSetPath(), 0700)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	return err
}

func (r *RuleSetLocation) getFullUrl() string {
	return "https://" + r.simpleUrl
}

func (r *RuleSetLocation) getRuleSetPath() string {
	dir, err := getRuleSetsDir()
	if err != nil {
		return ""
	}
	return dir + "/" + r.simpleUrl
}

func (r *RuleSetLocation) isRuleSetDownloaded() bool {
	_, err := os.ReadDir(r.getRuleSetPath())
	return errors.Is(err, fs.ErrNotExist)
}

func anyToError(val any) error {
	if err, ok := val.(error); ok {
		return err
	}
	if err, ok := val.(string); ok {
		return errors.New(err)
	}
	return fmt.Errorf("panic: %v", val)
}
