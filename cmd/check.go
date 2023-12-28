package cmd

import (
	"fmt"
	cmdApi "github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/cmd/guiApi"
	"github.com/avorty/spito/internal/checker"
	shared "github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/godbus/dbus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var checkFileCmd = &cobra.Command{
	Use:   "check file {path}",
	Short: "Check local lua rule file",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[1]

		runtimeData := getInitialRuntimeData(cmd)
		script, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed to read file %s\n", path)
			os.Exit(1)
		}

		doesRulePass, err := checker.CheckRuleScript(&runtimeData, string(script))
		if err != nil {
			panic(err)
		}

		communicateRuleResult(args[1], doesRulePass)
	},
}

var checkCmd = &cobra.Command{
	Use:   "check {ruleset identifier or path} {rule}",
	Short: "Check whether your machine pass rule",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runtimeData := getInitialRuntimeData(cmd)
		identifierOrPath := args[0]
		ruleName := args[1]

		if executionPath, err := os.Getwd(); err == nil {
			localRulesetPath := identifierOrPath
			if filepath.IsLocal(identifierOrPath) {
				localRulesetPath = filepath.Join(executionPath, identifierOrPath)
			}

			pathExists, err := shared.DoesPathExists(localRulesetPath)
			if err == nil && pathExists {
				identifierOrPath = localRulesetPath
			}
		}

		doesRulePass, err := checker.CheckRuleByIdentifier(&runtimeData, identifierOrPath, ruleName)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s", err)
			os.Exit(1)
		}

		communicateRuleResult(ruleName, doesRulePass)
	},
}

func getInitialRuntimeData(cmd *cobra.Command) shared.ImportLoopData {
	isExecutedByGui, err := cmd.Flags().GetBool("gui-child-mode")
	if err != nil {
		isExecutedByGui = true
	}

	var infoApi shared.InfoInterface

	if isExecutedByGui {
		conn, err := dbus.SessionBus()
		if err != nil {
			panic(err)
		}

		busObject := conn.Object("org.spito.gui", "/org/spito/gui")
		infoApi = guiApi.InfoApi{
			BusObject: busObject,
		}
	} else {
		infoApi = cmdApi.InfoApi{}
	}

	ruleVRCT, err := vrct.NewRuleVRCT()
	if err != nil {
		panic(err)
	}

	return shared.ImportLoopData{
		VRCT:         *ruleVRCT,
		RulesHistory: shared.RulesHistory{},
		ErrChan:      make(chan error),
		InfoApi:      infoApi,
	}
}

func communicateRuleResult(ruleName string, doesRulePass bool) {
	if doesRulePass {
		fmt.Printf("Rule %s successfuly passed requirements\n", ruleName)
	} else {
		fmt.Printf("Rule %s did not pass requirements\n", ruleName)
	}
}
