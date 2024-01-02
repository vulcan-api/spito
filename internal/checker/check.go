package checker

import (
	"errors"
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"os"
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
	return fmt.Errorf("panic: %v", val)
}


func CheckRuleByIdentifier(importLoopData *shared.ImportLoopData, identifier string, ruleName string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		return _internalCheckRule(importLoopData, identifier, ruleName, nil), nil
	})
}

func CheckRuleScript(importLoopData *shared.ImportLoopData, script string, rulesetPath string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		// TODO: implement preprocessing instead of hard coding ruleConf
		ruleConf := RuleConf{
			Path:   "",
			Unsafe: false,
		}
		script = processScript(script, &ruleConf, rulesetPath)
		return ExecuteLuaMain(script, importLoopData, &ruleConf)
	})
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
			if errChan != nil {
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
	previousRuleConf *RuleConf,
) bool {
	rulesetLocation := NewRulesetLocation(identifierOrPath)
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

	lockfilePath := rulesetLocation.GetRulesetPath() + "/" + LockFilename
	_, lockfileErr := os.ReadFile(lockfilePath)

	if os.IsNotExist(lockfileErr) {
		_, err := rulesetLocation.createLockfile(map[string]bool{})
		if err != nil {
			errChan <- errors.New("Failed to create dependency tree for: " + identifier + "\n" + err.Error())
			panic(nil)
		}
	}

	script, err := getScript(&rulesetLocation, ruleName)
	if err != nil {
		errChan <- errors.New("Failed to read script called: " + ruleName + " from " + identifier + "\n" + err.Error())
		panic(nil)
	}


	rulesetConf, err := getRulesetConf(&rulesetLocation)
	if err != nil {
		errChan <- fmt.Errorf("Failed to read %s config in git: %s \n%s", ConfigFilename, *rulesetLocation.GetFullUrl(), err.Error())
		panic(nil)
	}

	ruleConf := rulesetConf.Rules[ruleName]
	script = processScript(script, &ruleConf, rulesetLocation.GetRulesetPath())

	if previousRuleConf != nil {
		if !previousRuleConf.Unsafe && ruleConf.Unsafe {
			errChan <- errors.New("unsafe rule cannot be imported by safe rule")
			panic(nil)
		}
	}

	rulesHistory.SetProgress(identifier, ruleName, false)
	doesRulePass, err := ExecuteLuaMain(script, importLoopData, &ruleConf)
	if err != nil {
		return false
	}
	return doesRulePass
}
