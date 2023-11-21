package cmd

import (
	"fmt"
	"github.com/nasz-elektryk/spito/api"
	"github.com/nasz-elektryk/spito/checker"
	cmdApi "github.com/nasz-elektryk/spito/cmd/cmdApi"
	"github.com/nasz-elektryk/spito/cmd/guiApi"
	"github.com/spf13/cobra"
	"os"
)

var checkCmd = &cobra.Command{
	Use:   "check {ruleset identifier} {rule}",
	Short: "Check whether your machine pass rule",
	Args:  cobra.ExactArgs(2),
	Run:   checkRule,
}

func checkRule(cmd *cobra.Command, args []string) {
	identifier := args[0]
	ruleName := args[1]
	isExecutedByGui, err := cmd.Flags().GetBool("gui-child-mode")
	if err != nil {
		isExecutedByGui = true
	}

	var infoApi api.InfoInterface

	if isExecutedByGui {
		infoApi = guiApi.InfoApi{}
	} else {
		infoApi = cmdApi.InfoApi{}
	}

	runtimeData := checker.RuntimeData{
		RulesHistory: checker.RulesHistory{},
		ErrChan:      make(chan error),
		InfoApi:      infoApi,
	}

	doesRulePass, err := checker.CheckRuleByIdentifier(&runtimeData, identifier, ruleName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}

	if doesRulePass {
		fmt.Printf("Rule %s successfuly passed requirements\n", ruleName)
	} else {
		fmt.Printf("Rule %s did not pass requirements\n", ruleName)
	}
}
