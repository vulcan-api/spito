package checker

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
)

type RuleConf struct {
	Path string `yaml:"path"`
}

type SpitoRulesYaml struct {
	Rules map[string]interface{} `yaml:"rules"`
}

func (s SpitoRulesYaml) getRulesStringVal(key string) (string, bool) {
	val, ok := s.Rules[key].(string)
	return val, ok
}

func (s SpitoRulesYaml) getRulesStructVal(key string) (RuleConf, bool) {
	val, ok := s.Rules[key].(RuleConf)
	return val, ok
}

func getRulePath(rulesetPath, ruleName string) (string, error) {
	// Support for both .yaml and .yml
	spitoRulesDataBytes, err := os.ReadFile(rulesetPath + "/spito-rules.yaml")
	if errors.Is(err, fs.ErrNotExist) {
		var _err error
		spitoRulesDataBytes, _err = os.ReadFile(rulesetPath + "/spito-rules.yml")
		if _err != nil {
			return "", _err
		}
	} else if err != nil {
		return "", err
	}

	var spitoRulesYaml SpitoRulesYaml
	if err := yaml.Unmarshal(spitoRulesDataBytes, &spitoRulesYaml); err != nil {
		return "", err
	}

	for key := range spitoRulesYaml.Rules {
		if key != ruleName {
			continue
		}
		var path string

		if val, ok := spitoRulesYaml.getRulesStringVal(key); ok {
			path = val
		} else if val, ok := spitoRulesYaml.getRulesStructVal(key); ok {
			path = val.Path
		} else {
			continue
		}

		if path[0:2] == "./" {
			path = path[1:]
		} else if path[0] != '/' {
			path = "/" + path
		}
		path = rulesetPath + path

		return path, nil
	}

	return "", fmt.Errorf("NOT FOUND rule called: " + ruleName)
}

func getScript(rulesetPath, ruleName string) (string, error) {
	scriptPath, err := getRulePath(rulesetPath, ruleName)
	if err != nil {
		return "", err
	}

	script, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", err
	}
	return string(script), nil
}

func GetAllDownloadedRuleSets() ([]string, error) {
	ruleSetsDir, err := getRuleSetsDir()
	if err != nil {
		return nil, err
	}

	_ = initRequiredTmpDirs() // Ignore error because it should potentially avoid errors, not cause
	providerDirs, err := os.ReadDir(ruleSetsDir)
	if err != nil {
		return nil, err
	}
	var rulesets []string

	for _, provider := range providerDirs {
		providerName := provider.Name()
		userDirs, err := os.ReadDir(ruleSetsDir + "/" + providerName)
		if err != nil {
			continue
		}
		for _, user := range userDirs {
			userName := user.Name()
			userDirs, err := os.ReadDir(ruleSetsDir + "/" + providerName + "/" + userName)
			if err != nil {
				continue
			}
			for _, ruleSet := range userDirs {
				ruleSetName := ruleSet.Name()
				rulesets = append(rulesets,
					fmt.Sprintf("%s/%s/%s", providerName, userName, ruleSetName),
				)
			}
		}
	}

	return rulesets, nil
}

func FetchRuleset(rulesetLocation *RulesetLocation) error {
	err := rulesetLocation.CreateDir()
	if err != nil {
		println(err.Error())
		return err
	}

	_, err = git.PlainClone(rulesetLocation.GetRulesetPath(), false, &git.CloneOptions{
		URL: "https://" + *rulesetLocation.simpleUrl,
	})

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		println(err)
		return nil
	}
	return err
}
