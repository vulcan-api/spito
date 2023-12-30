package checker

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func getRulesetConf(ruleSetLocation RuleSetLocation) (RulesetConf, error) {
	// Support for both .yaml and .yml
	spitoRulesDataBytes, err := os.ReadFile(ruleSetLocation.GetRulesetPath() + "/spito-rules.yml")
	if os.IsNotExist(err) {
		var err2 error
		spitoRulesDataBytes, err2 = os.ReadFile(ruleSetLocation.GetRulesetPath() + "/spito-rules.yaml")
		if err2 != nil {
			return RulesetConf{}, err2
		}
	} else if err != nil {
		return RulesetConf{}, err
	}

	var spitoRulesYaml SpitoRulesetYaml
	if err := yaml.Unmarshal(spitoRulesDataBytes, &spitoRulesYaml); err != nil {
		return RulesetConf{}, err
	}

	spitoRulesetConf := RulesetConf{
		Rules: make(map[string]RuleConf),
	}

	for key := range spitoRulesYaml.Rules {
		ruleConf, err := spitoRulesYaml.getRuleConf(key)
		if err != nil {
			return RulesetConf{}, err
		}

		spitoRulesetConf.Rules[key] = ruleConf
	}

	return spitoRulesetConf, nil
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
	var ruleSets []string

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
				ruleSets = append(ruleSets,
					fmt.Sprintf("%s/%s/%s", providerName, userName, ruleSetName),
				)
			}
		}
	}

	return ruleSets, nil
}

type SpitoRulesetYaml struct {
	Rules map[string]interface{} `yaml:"rules"`
}

type RuleConfYaml struct {
	Path   string `yaml:"path"`
	Unsafe *bool  `yaml:"unsafe,omitempty"`
}

type RulesetConf struct {
	Rules map[string]RuleConf
}

type RuleConf struct {
	Path   string
	Unsafe bool
}

func (s SpitoRulesetYaml) getRuleConf(ruleName string) (RuleConf, error) {
	False := false // I did it in order to get pointer to false

	if ruleConfYaml, ok := s.Rules[ruleName].(RuleConfYaml); ok {
		ruleConfYaml.Path = filepath.Clean(ruleConfYaml.Path)
		if ruleConfYaml.Unsafe != nil {
			ruleConfYaml.Unsafe = &False
		}

		return RuleConf{
			Path:   ruleConfYaml.Path,
			Unsafe: *ruleConfYaml.Unsafe,
		}, nil
	}

	rulePath, ok := s.Rules[ruleName].(string)
	if !ok {
		return RuleConf{}, fmt.Errorf("rule %s in "+CONFIG_FILENAME+" is neither string nor RuleConfYaml", ruleName)
	}

	return RuleConf{
		Path:   filepath.Clean(rulePath),
		Unsafe: false,
	}, nil
}
