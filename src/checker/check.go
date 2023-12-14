package checker

import (
	"errors"
	"github.com/nasz-elektryk/spito/shared"
)

func CheckRuleByIdentifier(importLoopData *shared.ImportLoopData, identifier string, ruleName string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		return _internalCheckRule(importLoopData, identifier, ruleName), nil
	})
}

func CheckRuleScript(importLoopData *shared.ImportLoopData, script string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		return ExecuteLuaMain(script, importLoopData)
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
func _internalCheckRule(importLoopData *shared.ImportLoopData, identifier string, name string) bool {
	ruleSetLocation := RuleSetLocation{}
	ruleSetLocation.new(identifier)
	simpleUrl := ruleSetLocation.simpleUrl

	rulesHistory := &importLoopData.RulesHistory
	errChan := importLoopData.ErrChan

	if rulesHistory.Contains(simpleUrl, name) {
		if rulesHistory.IsRuleInProgress(simpleUrl, name) {
			errChan <- errors.New("ERROR: Dependencies creates infinity loop")
			panic(nil)
		} else {
			return true
		}
	}
	rulesHistory.Push(simpleUrl, name, true)

	err := FetchRuleSet(&ruleSetLocation)
	if err != nil {
		errChan <- errors.New("Failed to fetch rules from git: " + ruleSetLocation.getFullUrl() + "\n" + err.Error())
		panic(nil)
	}

	script, err := getScript(ruleSetLocation, name)
	if err != nil {
		errChan <- errors.New("Failed to read script called: " + name + " in git: " + ruleSetLocation.getFullUrl())
		panic(nil)
	}

	rulesHistory.SetProgress(simpleUrl, name, false)
	doesRulePass, err := ExecuteLuaMain(script, importLoopData)
	if err != nil {
		return false
	}
	return doesRulePass
}
