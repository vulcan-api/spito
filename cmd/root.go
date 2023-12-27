package cmd

import (
	"fmt"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/spf13/cobra"
	"os"
)

func handleError(errorToBePrinted error) {
	if errorToBePrinted != nil {
		printErrorAndExit(errorToBePrinted)
	}
}

func printErrorAndExit(errorToBePrinted error) {
	var infoApi cmdApi.InfoApi
	infoApi.Error(errorToBePrinted.Error())
	os.Exit(1)
}

type ConfigFileLayout struct {
	Repo_url   string
	Git_prefix string
	Identifier string
	Rules      map[string]string
}

var rootCmd = &cobra.Command{
	Use:   "spito",
	Short: "spito is powerful config management system",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Try running subcommand instead")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(checkFileCmd)
	rootCmd.AddCommand(newRulesetCommand)
	rootCmd.AddCommand(generateRuleCommand)
	rootCmd.AddCommand(generateShortCommand)

	checkFileCmd.Flags().Bool("gui-child-mode", false, "Tells app that it is executed by gui")
	checkCmd.Flags().Bool("gui-child-mode", false, "Tells app that it is executed by gui")
	newRulesetCommand.Flags().BoolP("non-interactive", "y", false, "If true assume default values for spito-rules.yaml")
}
