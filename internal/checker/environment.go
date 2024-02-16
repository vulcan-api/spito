package checker

import (
	"encoding/json"
	"errors"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
	"path/filepath"
)

var EnvironmentDataPath = filepath.Join(shared.LocalStateSpitoPath, "environment-data.json")
var NotEnvironmentErr = errors.New("called rule is not an environment")

type AppliedEnvironments []*AppliedEnvironment
type AppliedEnvironment struct {
	RevertNum        int    `json:"revertNumber"`
	IdentifierOrPath string `json:"identifierOrPath,omitempty"`
	IsApplied        bool
}

func ReadAppliedEnvironments() (AppliedEnvironments, error) {
	if err := shared.CreateIfNotExists(EnvironmentDataPath, "[]"); err != nil {
		return nil, err
	}

	environmentDataRaw, err := os.ReadFile(EnvironmentDataPath)
	if err != nil {
		return nil, err
	}

	appliedEnvironments := AppliedEnvironments{}
	err = json.Unmarshal(environmentDataRaw, &appliedEnvironments)

	return appliedEnvironments, err
}

func (e *AppliedEnvironments) Save() error {
	newContent, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return os.WriteFile(EnvironmentDataPath, newContent, shared.FilePermissions)
}

func (e *AppliedEnvironments) SetAsApplied(envIdentifierOrPath string, revertNum int) {
	foundEnv := false

	for _, env := range *e {
		if env.IsApplied && env.IdentifierOrPath != envIdentifierOrPath {
			env.IsApplied = false
		}
		if env.IdentifierOrPath == envIdentifierOrPath {
			foundEnv = true
			env.IsApplied = false
			env.RevertNum = revertNum
		}
	}
	if foundEnv {
		return
	}

	*e = append(*e, &AppliedEnvironment{
		RevertNum:        revertNum,
		IdentifierOrPath: envIdentifierOrPath,
		IsApplied:        true,
	})
}

func (e *AppliedEnvironments) RevertOther(envIdentifierOrPath string) error {
	for _, env := range *e {
		if env.IdentifierOrPath == envIdentifierOrPath || !env.IsApplied {
			continue
		}
		revertSteps, err := vrctFs.NewRevertSteps()
		if err != nil {
			return err
		}

		if err := revertSteps.Deserialize(env.RevertNum); err != nil {
			return err
		}

		if err := revertSteps.Apply(); err != nil {
			return err
		}

		env.IsApplied = false
	}

	return nil
}

func ApplyEnvironmentByIdentifier(importLoopData *shared.ImportLoopData, identifierOrPath string, envName string) error {
	doesEnvPass, err := checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		rulesetLocation := NewRulesetLocation(identifierOrPath)
		ruleConf, err := GetRuleConf(&rulesetLocation, envName)
		if err != nil {
			return false, err
		}

		if !ruleConf.Environment {
			return false, NotEnvironmentErr
		}
		return _internalCheckRule(importLoopData, identifierOrPath, envName, nil), nil
	})
	if err != nil {
		return err
	}
	if !doesEnvPass {
		return errors.New("environment didn't passed, cannot apply")
	}

	return applyEnvironment(importLoopData, identifierOrPath)
}

func ApplyEnvironmentScript(importLoopData *shared.ImportLoopData, script string, scriptPath string) error {
	doesEnvPass, err := checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		ruleConf := RuleConf{
			Path:   "",
			Unsafe: false,
		}
		script = processScript(script, &ruleConf)
		if !ruleConf.Environment {
			return false, NotEnvironmentErr
		}

		return ExecuteLuaMain(script, importLoopData, &ruleConf, filepath.Dir(scriptPath))
	})
	if err != nil {
		return err
	}
	if !doesEnvPass {
		return errors.New("the environment has not passed, cannot apply")
	}

	return applyEnvironment(importLoopData, scriptPath)
}

func applyEnvironment(importLoopData *shared.ImportLoopData, identifierOrPath string) error {
	appliedEnvironments, err := ReadAppliedEnvironments()
	if err != nil {
		return err
	}

	if err := appliedEnvironments.RevertOther(identifierOrPath); err != nil {
		return err
	}

	revertNum, err := importLoopData.VRCT.Apply()
	if err != nil {
		return err
	}

	appliedEnvironments.SetAsApplied(identifierOrPath, revertNum)
	return appliedEnvironments.Save()
}
