package checker

import (
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

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

func ReadSpitoYaml(rulesetLocation *RulesetLocation) ([]byte, error) {
	spitoRulesDataBytes, err := os.ReadFile(
		filepath.Join(rulesetLocation.GetRulesetPath(), shared.ConfigFilename))
	return spitoRulesDataBytes, err
}

func GetRulesetConf(rulesetLocation *RulesetLocation) (shared.ConfigFileLayout, error) {
	spitoRulesetYamlRaw, err := ReadSpitoYaml(rulesetLocation)
	if err != nil {
		return shared.ConfigFileLayout{}, err
	}

	var spitoRulesetYaml shared.ConfigFileLayout
	if err := yaml.Unmarshal(spitoRulesetYamlRaw, &spitoRulesetYaml); err != nil {
		return shared.ConfigFileLayout{}, err
	}

	return spitoRulesetYaml, nil
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

	// Get data from decorators
	processScript(string(scriptRaw), &ruleConf)

type RuleConfYaml struct {
	Path        string `yaml:"path"`
	Unsafe      bool   `yaml:"unsafe,omitempty"`
	Environment bool   `yaml:"environment,omitempty"`
	Sudo        bool   `yaml:"sudo,omitempty"`
}

type RulesetConf struct {
	Rules map[string]RuleConf
}

type RuleConf struct {
	Path        string
	Unsafe      bool
	Environment bool
	Sudo        bool
}
