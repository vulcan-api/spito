package shared

import (
	"errors"
	"fmt"
	"path/filepath"
)

const ConfigFilename = "spito.yml"
const LockFilename = "spito-lock.yml"

type RuleConfigLayout struct {
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Unsafe      bool   `yaml:"unsafe"`
	Environment bool
	Sudo        bool
}

type ConfigFileLayout struct {
	RepoUrl      string `yaml:"repo_url"`
	GitPrefix    string `yaml:"git_prefix"`
	Identifier   string
	Rules        map[string]RuleConfigLayout
	Description  string
	Branch       string
	Dependencies map[string][]string
}

func (s ConfigFileLayout) GetRuleConf(ruleName string) (RuleConfigLayout, error) {

	if ruleConfYaml, ok := s.Rules[ruleName]; ok {
		ruleConfYaml.Path = filepath.Clean(ruleConfYaml.Path)

		return RuleConfigLayout{
			Path:        ruleConfYaml.Path,
			Unsafe:      ruleConfYaml.Unsafe,
			Description: ruleConfYaml.Description,
		}, nil
	}
	return RuleConfigLayout{}, errors.New(fmt.Sprintf("cannot find rule named: '%s' in the config file", ruleName))
}
