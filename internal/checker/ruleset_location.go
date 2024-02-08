package checker

import (
	"errors"
	"github.com/avorty/spito/pkg/shared"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const ConfigFilename = "spito.yml"
const LockFilename = "spito-lock.yml"

func getRuleSetsDir() (string, error) {
	dir, err := os.UserHomeDir()
	return filepath.Join(dir, shared.LocalStateSpitoPath, "rulesets"), err
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

func GetDefaultRepoPrefix() string {
	// TODO: implement logic for getting default repo prefix
	return "github.com"
}

// RulesetLocation represent enum with value, only one of fields must be set
type RulesetLocation struct {
	simpleUrlOrPath string
	IsPath          bool
}

type DependencyTreeLayout struct {
	Dependencies []string
}

// e.g. from: https://github.com/avorty/spito-ruleset.git to avorty/spito-ruleset
func NewRulesetLocation(identifierOrPath string) RulesetLocation {
	r := RulesetLocation{}
	r.IsPath = false

	if filepath.IsAbs(identifierOrPath) {
		r.IsPath = true
		r.simpleUrlOrPath = identifierOrPath
		return r
	}

	// check if simpleUrlOrPath is url:
	if !strings.Contains(identifierOrPath, ".") {
		simpleUrl := GetDefaultRepoPrefix() + "/" + identifierOrPath
		simpleUrl = strings.ToLower(simpleUrl)

		r.simpleUrlOrPath = simpleUrl
		return r
	}

	simpleUrl := identifierOrPath
	simpleUrl = strings.ReplaceAll(simpleUrl, "https://", "")
	simpleUrl = strings.ReplaceAll(simpleUrl, "http://", "")
	simpleUrl = strings.ReplaceAll(simpleUrl, "www.", "")
	urlLen := len(simpleUrl)

	if simpleUrl[urlLen-1] == '/' {
		simpleUrl = simpleUrl[:urlLen-1]
	}
	// I still wonder whether it is good idea:
	if simpleUrl[urlLen-4:] == ".git" {
		simpleUrl = simpleUrl[:urlLen-4]
	}

	r.simpleUrlOrPath = strings.ToLower(simpleUrl)
	return r
}

func (r *RulesetLocation) GetIdentifier() string {
	return r.simpleUrlOrPath
}

func (r *RulesetLocation) GetFullUrl() *string {
	if r.IsPath {
		return nil
	}
	fullUrl := "https://" + r.simpleUrlOrPath
	return &fullUrl
}

func (r *RulesetLocation) CreateDir() error {
	err := os.MkdirAll(r.GetRulesetPath(), 0700)
	if errors.Is(err, fs.ErrExist) {
		return nil
	}
	return err
}

func (r *RulesetLocation) GetRulesetPath() string {
	if r.IsPath {
		return r.simpleUrlOrPath
	}

	dir, err := getRuleSetsDir()
	if err != nil {
		return ""
	}
	return dir + "/" + r.simpleUrlOrPath
}

func (r *RulesetLocation) IsRuleSetDownloaded() bool {
	_, err := os.ReadDir(r.GetRulesetPath())
	return !errors.Is(err, fs.ErrNotExist)
}

func (r *RulesetLocation) createLockfile(rulesInProgress map[string]bool) ([]string, error) {
	configFileContents, err := ReadRawSpitoYaml(r)
	if err != nil {
		return []string{}, err
	}

	var basicDependencyTree DependencyTreeLayout

	err = yaml.Unmarshal(configFileContents, &basicDependencyTree)
	if err != nil {
		return []string{}, err
	}

	var outputDependencyTree DependencyTreeLayout
	outputDependencyTree.Dependencies = make([]string, len(basicDependencyTree.Dependencies))
	copy(outputDependencyTree.Dependencies, basicDependencyTree.Dependencies)

	var firstDependencyLocation RulesetLocation

	if len(basicDependencyTree.Dependencies) > 0 {
		firstDependencyLocation = NewRulesetLocation(strings.Split(basicDependencyTree.Dependencies[0], "@")[0])
	}

	var waitGroup sync.WaitGroup

	for _, dependencyName := range basicDependencyTree.Dependencies {
		waitGroup.Add(1)
		go func(dependencyNameParameter string) {
			defer waitGroup.Done()
			dependencyLocation := NewRulesetLocation(strings.Split(dependencyNameParameter, "@")[0])
			if _, exists := rulesInProgress[dependencyNameParameter]; !exists && !dependencyLocation.IsRuleSetDownloaded() {
				rulesInProgress[dependencyNameParameter] = true
				FetchRuleset(&dependencyLocation)
			}
		}(dependencyName)
	}

	waitGroup.Wait()
	if firstDependencyLocation.simpleUrlOrPath != "" && !firstDependencyLocation.IsPath {
		toBeAppended, err := firstDependencyLocation.createLockfile(rulesInProgress)
		if err != nil {
			return nil, err
		}
		outputDependencyTree.Dependencies = append(outputDependencyTree.Dependencies, toBeAppended...)
	}

	lockfilePath := r.GetRulesetPath() + "/" + LockFilename
	lockfile, err := os.Create(lockfilePath)

	if err != nil {
		return []string{}, err
	}
	defer lockfile.Close()

	yamlOutput, err := yaml.Marshal(outputDependencyTree)
	if err != nil {
		return []string{}, err
	}

	lockfile.Write(yamlOutput)

	return outputDependencyTree.Dependencies, nil
}
