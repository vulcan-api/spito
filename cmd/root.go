package cmd

import (
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
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
	Path        string `yaml:"path"`
	Description string `yaml:"description"`
	Unsafe      bool   `yaml:"unsafe"`
}

type ConfigFileLayout struct {
	RepoUrl     string `yaml:"repo_url"`
	GitPrefix   string `yaml:"git_prefix"`
	Identifier  string
	Rules       map[string]Rule
	Description string
	Branch      string
}

func readConfigFile(rulesetPath string, output *ConfigFileLayout) {
	configFileContents, err := os.ReadFile(filepath.Join(rulesetPath, checker.ConfigFilename))
	handleError(err)

	err = yaml.Unmarshal(configFileContents, &output)
	handleError(err)
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
	rootCmd.AddCommand(revertCmd)
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
