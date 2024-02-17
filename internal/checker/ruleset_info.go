package checker

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
)

func GetRulesetConf(rulesetLocation *RulesetLocation) (RulesetConf, error) {
	spitoRulesetConf := RulesetConf{
		Rules: make(map[string]RuleConf),
	}

	rulesetConfYaml, err := getRulesetConfYaml(rulesetLocation)
	if err != nil {
		return RulesetConf{}, err
	}

	for key := range rulesetConfYaml.Rules {
		ruleConf, err := rulesetConfYaml.GetRuleConfBasedOnYaml(rulesetLocation, key)
		if err != nil {
			return RulesetConf{}, err
		}

		spitoRulesetConf.Rules[key] = ruleConf
	}

	return spitoRulesetConf, nil
}

func GetRuleConf(rulesetLocation *RulesetLocation, ruleName string) (RuleConf, error) {
	rulesetConfYaml, err := getRulesetConfYaml(rulesetLocation)
	if err != nil {
		return RuleConf{}, err
	}

	return rulesetConfYaml.GetRuleConfBasedOnYaml(rulesetLocation, ruleName)
}

func GetRuleConfFromScript(scriptPath string) (RuleConf, error) {
	ruleConf := RuleConf{
		Path:        scriptPath,
		Unsafe:      false,
		Environment: false,
	}

	scriptRaw, err := os.ReadFile(scriptPath)
	if err != nil {
		return RuleConf{}, err
	}

	// Get data from decorators
	processScript(string(scriptRaw), &ruleConf)

	return ruleConf, err
}

func (s *RulesetConfYaml) GetRuleConfBasedOnYaml(rulesetLocation *RulesetLocation, ruleName string) (RuleConf, error) {
	var ruleConf RuleConf

	if ruleConfYaml, ok := s.Rules[ruleName].(RuleConfYaml); ok {
		ruleConfYaml.Path = filepath.Clean(ruleConfYaml.Path)

		ruleConf = RuleConf{
			Path:        ruleConfYaml.Path,
			Unsafe:      ruleConfYaml.Unsafe,
			Environment: ruleConfYaml.Environment,
		}
	} else if rulePath, ok := s.Rules[ruleName].(string); ok {
		ruleConf = RuleConf{
			Path:   filepath.Clean(rulePath),
			Unsafe: false,
		}
	} else {
		return RuleConf{}, fmt.Errorf("rule %s in %s is neither string nor RuleConfYaml", ruleName, ConfigFilename)
	}

	scriptPath := filepath.Join(rulesetLocation.GetRulesetPath(), ruleConf.Path)
	scriptRaw, err := os.ReadFile(scriptPath)
	if err != nil {
		return RuleConf{}, err
	}

	// Get data from decorators
	processScript(string(scriptRaw), &ruleConf)

	return ruleConf, err
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

func getRulesetConfYaml(rulesetLocation *RulesetLocation) (RulesetConfYaml, error) {
	if err := FetchRuleset(rulesetLocation); err != nil {
		return RulesetConfYaml{}, err
	}

	rawSpitoYaml, err := ReadRawSpitoYaml(rulesetLocation)
	if err != nil {
		return RulesetConfYaml{}, err
	}

	var spitoRulesYaml RulesetConfYaml
	err = yaml.Unmarshal(rawSpitoYaml, &spitoRulesYaml)

	return spitoRulesYaml, err
}

func ReadRawSpitoYaml(rulesetLocation *RulesetLocation) ([]byte, error) {
	spitoYamlPath := path.Join(rulesetLocation.GetRulesetPath(), "spito.yml")
	spitoRulesDataBytes, err := os.ReadFile(spitoYamlPath)
	
	if os.IsNotExist(err) {
		spitoYamlPath := path.Join(rulesetLocation.GetRulesetPath(), "spito.yaml")
		spitoRulesDataBytes, err = os.ReadFile(spitoYamlPath)
		if err != nil {
			return nil, err
		}
	}
	return spitoRulesDataBytes, err
}

type RulesetConfYaml struct {
	Rules map[string]interface{} `yaml:"rules"`
}

type RuleConfYaml struct {
	Path        string `yaml:"path"`
	Unsafe      bool   `yaml:"unsafe,omitempty"`
	Environment bool   `yaml:"environment,omitempty"`
}

type RulesetConf struct {
	Rules map[string]RuleConf
}

type RuleConf struct {
	Path        string
	Unsafe      bool
	Environment bool
}
