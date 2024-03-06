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

func FetchRuleset(rulesetLocation *RulesetLocation) error {
	err := rulesetLocation.CreateDir()
	if err != nil {
		return err
	}

	fullRulesetUrl := *rulesetLocation.GetFullUrl()

	_, err = git.PlainClone(rulesetLocation.GetRulesetPath(), false, &git.CloneOptions{
		URL: fullRulesetUrl,
	})

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		repo, err := git.PlainOpen(rulesetLocation.GetRulesetPath())
		if err != nil {
			return err
		}

		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}

		// We force pull because nobody should modify by themselves rulesets in their spito directory
		err = worktree.Pull(&git.PullOptions{Force: true, RemoteURL: fullRulesetUrl})
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}

		return err
	}
	return err
}
