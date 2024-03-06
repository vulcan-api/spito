package checker

import (
	"errors"
	"github.com/avorty/spito/pkg/path"
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

		isPath, err := path.PathExists(rule.Url)
		if err != nil || rule.Url == "" {
			isPath = false
		}

		rulesetLocation, err := NewRulesetLocation(rule.Url, isPath)
		if err != nil {
			return err
		}

		var script string
		if rule.IsScript {
			script = rule.NameOrScript
		} else {
			script, err = getScript(&rulesetLocation, rule.NameOrScript)
			if err != nil {
				return err
			}
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		ruleConfigLayout := shared.RuleConfigLayout{Path: cwd}
		script = processScript(script, &ruleConfigLayout)

		// TODO: Passing here cwd is not the best idea
		L, err := GetLuaState(script, &importLoopData, &ruleConfigLayout, cwd)
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
