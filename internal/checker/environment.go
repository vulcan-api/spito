package checker

import (
	"encoding/json"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
	"path/filepath"
)

var EnvironmentDataPath = filepath.Join(shared.UserHomeDir, ".local/state/spito/environment-data.json")

type AppliedEnvironments []*AppliedEnvironment
type AppliedEnvironment struct {
	RevertNum        int    `json:"revertNumber"`
	IdentifierOrPath string `json:"identifierOrPath,omitempty"`
	IsApplied        bool
}

func ReadAppliedEnvironments() (AppliedEnvironments, error) {
	if err := shared.CreateIfNotExist(EnvironmentDataPath, "[]"); err != nil {
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
	return os.WriteFile(EnvironmentDataPath, newContent, os.ModePerm)
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
	}

	return nil
}

func ApplyEnvironmentByIdentifier(importLoopData *shared.ImportLoopData, identifier string, ruleName string) (bool, error) {
	doesEnvPass, err := checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		return _internalCheckRule(importLoopData, identifier, ruleName, nil), nil
	})
	if err != nil {
		return false, err
	}
	if !doesEnvPass {
		return doesEnvPass, err
	}

	appliedEnvironments, err := ReadAppliedEnvironments()
	if err != nil {
		return false, err
	}
	if err := appliedEnvironments.RevertOther(identifier); err != nil {
		return false, err
	}

	revertNum, err := importLoopData.VRCT.Apply()
	if err != nil {
		return false, err
	}

	appliedEnvironments.SetAsApplied(identifier, revertNum)
	return doesEnvPass, appliedEnvironments.Save()
}

func ApplyEnvironmentScript(importLoopData *shared.ImportLoopData, script string, scriptPath string) (bool, error) {
	doesEnvPass, err := checkAndProcessPanics(importLoopData, func(errChan chan error) (bool, error) {
		ruleConf := RuleConf{
			Path:   "",
			Unsafe: false,
		}
		script = processScript(script, &ruleConf)
		return ExecuteLuaMain(script, importLoopData, &ruleConf, filepath.Dir(scriptPath))
	})
	if err != nil {
		return false, err
	}
	if !doesEnvPass {
		return doesEnvPass, err
	}

	appliedEnvironments, err := ReadAppliedEnvironments()
	if err != nil {
		return false, err
	}
	if err := appliedEnvironments.RevertOther(scriptPath); err != nil {
		return false, err
	}

	revertNum, err := importLoopData.VRCT.Apply()
	if err != nil {
		return false, err
	}

	appliedEnvironments.SetAsApplied(scriptPath, revertNum)
	return doesEnvPass, appliedEnvironments.Save()
}
