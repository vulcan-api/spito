package checker

import (
	"errors"
	"github.com/avorty/spito/pkg/path"
	"github.com/avorty/spito/pkg/shared"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func getRuleSetsDir() string {
	return filepath.Join(shared.LocalStateSpitoPath, "rulesets")
}

func initRequiredTmpDirs() error {
	dir := getRuleSetsDir()

	err := os.MkdirAll(dir, 0700)
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
	Dependencies map[string][]string `yaml:"dependencies"`
}

// NewRulesetLocation e.g., from: https://github.com/avorty/spito-ruleset.git to avorty/spito-ruleset
func NewRulesetLocation(identifierOrPath string, isPath bool) (RulesetLocation, error) {
	r := RulesetLocation{}
	r.IsPath = isPath

	if isPath {
		absolutePath, err := filepath.Abs(identifierOrPath)
		if err != nil {
			return RulesetLocation{}, err
		}
		r.simpleUrlOrPath = absolutePath
		return r, nil
	}

	// check if simpleUrlOrPath is url:
	if !strings.Contains(identifierOrPath, ".") {
		simpleUrl := GetDefaultRepoPrefix() + "/" + identifierOrPath
		simpleUrl = strings.ToLower(simpleUrl)

		r.simpleUrlOrPath = simpleUrl

		err := FetchRuleset(&r)
		return r, err
	}

	simpleUrl := identifierOrPath
	simpleUrl = strings.ReplaceAll(simpleUrl, "https://", "")
	simpleUrl = strings.ReplaceAll(simpleUrl, "http://", "")
	simpleUrl = strings.ReplaceAll(simpleUrl, "www.", "")
	urlLen := len(simpleUrl)

	if simpleUrl[urlLen-1] == '/' {
		simpleUrl = simpleUrl[:urlLen-1]
	}
	// I still wonder whether it is a good idea:
	if simpleUrl[urlLen-4:] == ".git" {
		simpleUrl = simpleUrl[:urlLen-4]
	}

	r.simpleUrlOrPath = strings.ToLower(simpleUrl)

	err := FetchRuleset(&r)
	return r, err
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

	return getRuleSetsDir()
}

func (r *RulesetLocation) IsRuleSetDownloaded() bool {
	_, err := os.ReadDir(r.GetRulesetPath())
	return !errors.Is(err, fs.ErrNotExist)
}

func InstallDependency(ruleIdentifier string, waitGroup *sync.WaitGroup, errChan chan error) {
	var err error
	defer waitGroup.Done()
	dependencyLocation, err := NewRulesetLocation(strings.Split(ruleIdentifier, "@")[0], false)
	if err != nil {
		errChan <- err
		panic(nil)
	}

	if !dependencyLocation.IsRuleSetDownloaded() {
		err = FetchRuleset(&dependencyLocation)
	}
	if err != nil {
		errChan <- err
		panic(nil)
	}
}

func (r *RulesetLocation) getLockfileTree() (DependencyTreeLayout, error) {
	fileContents, err := os.ReadFile(filepath.Join(r.GetRulesetPath(), shared.LockFilename))
	if err != nil {
		return DependencyTreeLayout{}, err
	}

	var output DependencyTreeLayout
	err = yaml.Unmarshal(fileContents, &output)
	if err != nil {
		return DependencyTreeLayout{}, err
	}
	return output, nil
}

func (r *RulesetLocation) createLockfile(errChan chan error) error {

	basicDependencyTree, err := GetRulesetConf(r)
	if err != nil {
		return err
	}

	var waitGroup sync.WaitGroup

	for _, ruleDependencies := range basicDependencyTree.Dependencies {
		for _, dependencyString := range ruleDependencies {
			waitGroup.Add(1)
			go InstallDependency(dependencyString, &waitGroup, errChan)
		}
	}
	waitGroup.Wait()

	for ruleName, ruleDependencies := range basicDependencyTree.Dependencies {
		for _, dependencyString := range ruleDependencies {
			dependencyRulesetName, dependencyRuleName, _ := strings.Cut(dependencyString, "@")
			rulesetLocation, err := NewRulesetLocation(dependencyRulesetName, false)
			if err != nil {
				return err
			}

			doesLockfileExist, err := path.PathExists(filepath.Join(rulesetLocation.GetRulesetPath(), shared.LockFilename))
			if err != nil {
				return err
			}
			if !doesLockfileExist {
				err = rulesetLocation.createLockfile(errChan)
			}
			if err != nil {
				return err
			}

			dependencyTree, err := rulesetLocation.getLockfileTree()
			if err != nil {
				return err
			}

			basicDependencyTree.Dependencies[ruleName] =
				append(basicDependencyTree.Dependencies[ruleName], dependencyTree.Dependencies[dependencyRuleName]...)
		}
	}

	lockfilePath := filepath.Join(r.GetRulesetPath(), shared.LockFilename)
	lockfile, err := os.Create(lockfilePath)

	if err != nil {
		return err
	}
	defer func() {
		err = lockfile.Close()
		if err != nil {
			errChan <- err
			panic(nil)
		}
	}()

	yamlOutput, err := yaml.Marshal(DependencyTreeLayout{
		Dependencies: basicDependencyTree.Dependencies,
	})
	if err != nil {
		return err
	}

	_, err = lockfile.Write(yamlOutput)
	if err != nil {
		return err
	}

	_, err = lockfile.Write(yamlOutput)
	if err != nil {
		return nil
	}

	return nil
}
