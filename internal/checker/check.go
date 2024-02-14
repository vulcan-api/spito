package checker

import (
	"errors"
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"os"
	"path/filepath"
	"strings"
)

type Rule struct {
	url          string
	name         string
	isInProgress bool
}

type RulesHistory map[string]Rule

func (r RulesHistory) Contains(url string, name string) bool {
	_, ok := r[url+name]
	return ok
}

func (r RulesHistory) IsRuleInProgress(url string, name string) bool {
	val := r[url+name]
	return val.isInProgress
}

func (r RulesHistory) Push(url string, name string, isInProgress bool) {
	r[url+name] = Rule{
		url:          url,
		name:         name,
		isInProgress: isInProgress,
	}
}

func (r RulesHistory) SetProgress(url string, name string, isInProgress bool) {
	rule := r[url+name]
	rule.isInProgress = isInProgress
}

func anyToError(val any) error {
	if err, ok := val.(error); ok {
		return err
	}
	if err, ok := val.(string); ok {
		return errors.New(err)
	}
	return fmt.Errorf("panic: %+v", val)
}

func CheckRuleByPath(importLoopData *shared.ImportLoopData, rulesetPath string, ruleName string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		return _internalCheckRule(importLoopData, rulesetPath, ruleName, nil, true), nil
	})
}

func CheckRuleByIdentifier(importLoopData *shared.ImportLoopData, identifier string, ruleName string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		return _internalCheckRule(importLoopData, identifier, ruleName, nil, false), nil
	})
}

func CheckRuleScript(importLoopData *shared.ImportLoopData, script string, scriptDirectory string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		// TODO: implement preprocessing instead of hard coding ruleConf
		ruleConf := shared.RuleConfigLayout{}
		script, err := processScript(script, &ruleConf)
		if err != nil {
			return false, err
		}
		return ExecuteLuaMain(script, importLoopData, &ruleConf, scriptDirectory)
	})
}

func handleErrorAndPanic(errChan chan error, err error) {
	if err != nil {
		errChan <- err
		panic(nil)
	}
}

func checkAndProcessPanics(
	importLoopData *shared.ImportLoopData,
	checkFunc func(errChan chan error) (bool, error),
) (bool, error) {

	errChan := importLoopData.ErrChan
	doesRulePassChan := make(chan bool)
	go func() {
		defer func() {
			r := recover()
			if errChan != nil && r != nil {
				errChan <- anyToError(r)
			}
		}()

		doesRulePass, err := checkFunc(errChan)
		if err != nil {
			errChan <- err
		}
		doesRulePassChan <- doesRulePass
	}()

	select {
	case doesRulePass := <-doesRulePassChan:
		return doesRulePass, nil
	case err := <-errChan:
		return false, err
	}
}

// This function shouldn't be executed directly,
// because in case of panic it does not handle errors at all
func _internalCheckRule(
	importLoopData *shared.ImportLoopData,
	identifierOrPath string,
	ruleName string,
	previousRuleConf *shared.RuleConfigLayout,
	isPath bool,
) bool {
	rulesetLocation := NewRulesetLocation(identifierOrPath, isPath)
	identifier := rulesetLocation.GetIdentifier()

	rulesHistory := &importLoopData.RulesHistory
	errChan := importLoopData.ErrChan

	if rulesHistory.Contains(identifier, ruleName) {
		if rulesHistory.IsRuleInProgress(identifier, ruleName) {
			errChan <- errors.New("ERROR: Dependencies creates infinity loop")
			panic(nil)
		} else {
			return true
		}
	}
	rulesHistory.Push(identifier, ruleName, true)

	if !rulesetLocation.IsPath {
		err := FetchRuleset(&rulesetLocation)
		if err != nil {
			errChan <- errors.New("Failed to fetch rules from: " + identifier + "\n" + err.Error())
			panic(nil)
		}
	}
	lockfilePath := filepath.Join(rulesetLocation.GetRulesetPath(), shared.LockFilename)
	_, lockfileErr := os.ReadFile(lockfilePath)

	if os.IsNotExist(lockfileErr) {
		err := rulesetLocation.createLockfile(importLoopData.ErrChan)
		if err != nil {
			errChan <- errors.New("Failed to create dependency tree for: " + identifier + "\n" + err.Error())
			panic(nil)
		}
	}

	dependencies, err := rulesetLocation.getLockfileTree()
	if err != nil {
		errChan <- err
		panic(nil)
	}

	for _, dependencyString := range dependencies.Dependencies[ruleName] {
		importLoopData.InfoApi.Log(fmt.Sprintf("Checking requirements for the dependency '%s'", dependencyString))
		rulesetName, ruleName, _ := strings.Cut(dependencyString, "@")
		doesDependencyPass := _internalCheckRule(importLoopData, rulesetName, ruleName, previousRuleConf, false)
		if !doesDependencyPass {
			errChan <- errors.New(fmt.Sprintf("Rule %s did not pass requirements", ruleName))
			return false
		}
	}

	script, err := getScript(&rulesetLocation, ruleName)
	if err != nil {
		errChan <- errors.New("Failed to read script called: " + ruleName + " from " + identifier + "\n" + err.Error() + "\n")
		panic(nil)
	}

	rulesetConf, err := GetRulesetConf(&rulesetLocation)
	if err != nil {
		errChan <- fmt.Errorf("Failed to read %s config in git: %s \n%s", shared.ConfigFilename, *rulesetLocation.GetFullUrl(), err.Error())
		panic(nil)
	}

	ruleConf := rulesetConf.Rules[ruleName]
	processedScript, err := processScript(script, &ruleConf)
	if err != nil {
		errChan <- fmt.Errorf("Failed to process script >>>\n%s<<< from %s: %s\n", script, ruleConf.Path, err.Error())
		panic(nil)
	}

	if previousRuleConf != nil {
		if !previousRuleConf.Unsafe && ruleConf.Unsafe {
			errChan <- errors.New("unsafe rule cannot be imported by safe rule")
			panic(nil)
		}
	}

	isRunAsRoot, err := shared.IsRoot()
	if err != nil {
		errChan <- err
		panic(nil)
	}

	if ruleConf.Sudo && !isRunAsRoot {
		errChan <- errors.New("tried to execute a spito rule that requires root privileges")
		panic(nil)
	}

	rulesHistory.SetProgress(identifier, ruleName, false)
	doesRulePass, err := ExecuteLuaMain(processedScript, importLoopData, &ruleConf, rulesetLocation.GetRulesetPath())
	if err != nil {
		errChan <- err
		panic(nil)
	}
	return doesRulePass
}
