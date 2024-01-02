package cmd

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"path/filepath"

	cmdApi "github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/cmd/guiApi"
	"github.com/avorty/spito/internal/checker"
	shared "github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/godbus/dbus"
	"github.com/spf13/cobra"
)

var checkFileCmd = &cobra.Command{
	Use:   "file {path}",
	Short: "Check local lua rule file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		runtimeData := getInitialRuntimeData(cmd)
		script, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed to read file %s\n", path)
			os.Exit(1)
		}

		path, err = filepath.Abs(path)
		if err != nil {
			infoApi := cmdApi.InfoApi{}
			infoApi.Error("Error during conversion to absolute path!")
			os.Exit(1)
		}
		
		directories := strings.Split(path, "/")
		rulesDirectoryIndex := slices.Index(directories, "rules")
		
		if len(directories) < 2 || rulesDirectoryIndex == -1 {
			runtimeData.InfoApi.Error("Your rule must be correctly placed inside a ruleset!")
			os.Exit(1)
		}
		
		rulesetPath := strings.Join(directories[:len(directories)-2], "/")
		_, err = os.Stat(rulesetPath + "/" + checker.ConfigFilename)

		if os.IsNotExist(err) {
			runtimeData.InfoApi.Error(fmt.Sprintf("There's no %s file in %s", checker.ConfigFilename, rulesetPath))
			os.Exit(1)
		}
		
		doesRulePass, err := checker.CheckRuleScript(&runtimeData, string(script), rulesetPath)
		if err != nil {
			panic(err)
		}

		communicateRuleResult(path, doesRulePass)
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

			pathExists, err := shared.DoesPathExist(localRulesetPath)
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
