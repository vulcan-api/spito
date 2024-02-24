package checker

import (
	"errors"
	"fmt"
	"github.com/avorty/spito/pkg/shared"
	"os"
	"os/user"
)

type Rule struct {
	Url          string `json:"Url" bson:"Url"`
	Name         string `json:"Name" bson:"Name"`
	IsInProgress bool   `json:"IsInProgress" bson:"IsInProgress"`
}

type RulesHistory map[string]Rule

func (r RulesHistory) Contains(url string, name string) bool {
	_, ok := r[url+name]
	return ok
}

func (r RulesHistory) IsRuleInProgress(url string, name string) bool {
	val := r[url+name]
	return val.IsInProgress
}

func (r RulesHistory) Push(url string, name string, isInProgress bool) {
	r[url+name] = Rule{
		Url:          url,
		Name:         name,
		IsInProgress: isInProgress,
	}
}

func (r RulesHistory) SetProgress(url string, name string, isInProgress bool) {
	rule := r[url+name]
	rule.IsInProgress = isInProgress
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

func CheckRuleByIdentifier(importLoopData *shared.ImportLoopData, identifier string, ruleName string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		return _internalCheckRule(importLoopData, identifier, ruleName, nil), nil
	})
}

func CheckRuleScript(importLoopData *shared.ImportLoopData, script string, scriptDirectory string) (bool, error) {
	return checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		ruleConf := RuleConf{}
		script = processScript(script, &ruleConf)

		L, err := GetLuaState(script, importLoopData, &ruleConf, scriptDirectory)
		if err != nil {
			return false, err
		}
		defer L.Close()

		return ExecuteLuaMain(L)
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

	rulesetConf, err := GetRulesetConf(&rulesetLocation)
	if err != nil {
		errChan <- fmt.Errorf("Failed to read %s config in git: %s \n%s", ConfigFilename, *rulesetLocation.GetFullUrl(), err.Error())
		panic(nil)
	}

	ruleConf := rulesetConf.Rules[ruleName]
	script = processScript(script, &ruleConf)

	if previousRuleConf != nil {
		if !previousRuleConf.Unsafe && ruleConf.Unsafe {
			errChan <- errors.New("unsafe rule cannot be imported by safe rule")
			panic(nil)
		}
	}

	isRunAsRoot, err := isRoot()
	if err != nil {
		errChan <- err
		panic(nil)
	}

	if ruleConf.Sudo && !isRunAsRoot {
		errChan <- errors.New("tried to execute a spito rule that requires root privileges")
		panic(nil)
	}

	rulesHistory.SetProgress(identifier, ruleName, false)

	L, err := GetLuaState(script, importLoopData, &ruleConf, rulesetLocation.GetRulesetPath())
	if err != nil {
		errChan <- err
		panic(nil)
	}

	doesRulePass, err := ExecuteLuaMain(L)
	if err != nil {
		errChan <- err
		panic(nil)
	}
	return doesRulePass
}

func isRoot() (bool, error) {
	currentUser, err := user.Current()
	return currentUser.Username == "root", err
}
