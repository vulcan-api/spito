package checker

import (
	"errors"
	"github.com/go-git/go-git/v5"
	"os"
	"path/filepath"
)

func getScript(ruleSetLocation *RulesetLocation, ruleName string) (string, error) {
	rulesetConf, err := GetRulesetConf(ruleSetLocation)
	if err != nil {
		return "", err
	}
	ruleConf := rulesetConf.Rules[ruleName]
	scriptPath := filepath.Join(ruleSetLocation.GetRulesetPath(), ruleConf.Path)
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}
	return string(script), nil
}

func FetchRuleset(ruleSetLocation *RulesetLocation) error {
	err := ruleSetLocation.CreateDir()
	if err != nil {
		println(err.Error())
		return err
	}

	_, err = git.PlainClone(ruleSetLocation.GetRulesetPath(), false, &git.CloneOptions{
		URL: *ruleSetLocation.GetFullUrl(),
	})

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return nil
	}
	return err
}
