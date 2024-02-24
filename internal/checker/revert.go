package checker

import (
	"errors"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"os"
)

func GetRevertRuleFn(infoApi shared.InfoInterface) func(rule vrctFs.Rule) error {
	return func(rule vrctFs.Rule) error {
		importLoopData := shared.ImportLoopData{
			VRCT:         vrct.RuleVRCT{},
			InfoApi:      infoApi,
			RulesHistory: make(shared.RulesHistory),
			ErrChan:      make(chan error),
		}

		isPath, err := shared.PathExists(rule.Url)
		if err != nil {
			isPath = false
		}

		rulesetLocation := NewRulesetLocation(rule.Url, isPath)

		script, err := getScript(&rulesetLocation, rule.Name)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// TODO: Passing here cwd is not the best idea
		L, err := GetLuaState(script, &importLoopData, &shared.RuleConfigLayout{}, cwd)
		if err != nil {
			return err
		}
		defer L.Close()

		pass, err := ExecuteLuaRevert(L)

		if err != nil {
			return err
		}
		if !pass {
			return errors.New("revert failed by returning false")
		}

		return nil
	}
}
