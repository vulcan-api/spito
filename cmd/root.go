package cmd

import (
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

type Rule struct {
	Path string `yaml:"path"`
	Description string `yaml:"description"`
	Unsafe bool `yaml:"unsafe"`
}

type ConfigFileLayout struct {
	Repo_url   string
	Git_prefix string
	Identifier string
	Rules      map[string]Rule
	Description string
}

var rootCmd = &cobra.Command{
	Use:   "spito",
	Short: "spito is powerful config management system",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			printErrorAndExit(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printErrorAndExit(err)
	}
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.AddCommand(checkFileCmd)
	rootCmd.AddCommand(newRulesetCommand)
	rootCmd.AddCommand(generateRuleCommand)
	rootCmd.AddCommand(generateShortCommand)
	rootCmd.AddCommand(loginCommand)
	rootCmd.AddCommand(publishCommand)

	checkFileCmd.Flags().Bool("gui-child-mode", false, "Tells app that it is executed by gui")
	checkCmd.Flags().Bool("gui-child-mode", false, "Tells app that it is executed by gui")
	newRulesetCommand.Flags().BoolP("non-interactive", "y", false, "If true assume default values for spito-rules.yaml")
	loginCommand.Flags().BoolP("local", "l", false, "If true, save login credentials inside a spito ruleset")
	publishCommand.Flags().BoolP("local", "l", false, "If true, get login token from a local ruleset")
}
