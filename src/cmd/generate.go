package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nasz-elektryk/spito/checker"
	"github.com/nasz-elektryk/spito/cmd/cmdApi"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const exampleRuleContents = `function main()
  return true
end
`

func handleGenerate(cmd *cobra.Command, args []string) {
	configFileContents, error := os.ReadFile(checker.CONFIG_FILENAME)
	if os.IsNotExist(error) {
		printErrorAndExit(errors.New("Please run this commnd inside an actual spito ruleset directory!"))
	}
	handleError(error)

	config := ConfigFileLayout{}
	error = yaml.Unmarshal(configFileContents, &config)
	handleError(error)

	rulesetLocation := checker.RuleSetLocation{}
	rulesetLocation.New(config.Identifier)
	if !rulesetLocation.IsRuleSetDownloaded() {
		printErrorAndExit(errors.New("Your ruleset is not in the standard ruleset path!"))
	}

	ruleName := strings.ReplaceAll(args[0], " ", "")
	ruleFile, error := os.Create("rules/" + ruleName + ".lua")
	handleError(error)
	ruleFile.Write([]byte(exampleRuleContents))
	ruleFile.Close()

	config.Rules[ruleName] = "./rules/" + ruleName + ".lua"

	configFile, error := os.Create(checker.CONFIG_FILENAME)
	defer configFile.Close()
	handleError(error)
	configFileContents, error = yaml.Marshal(config)
	handleError(error)
	configFile.Write(configFileContents)

	infoApi := cmdApi.InfoApi{}
	infoApi.Log(fmt.Sprintf("Successfully created rule '%s'", ruleName))
}

var generateRuleCommand = &cobra.Command{
	Use:   "generate {rule_name}",
	Short: "Generate new rule",
	Args:  cobra.ExactArgs(1),
	Run: handleGenerate,
}

var generateShortCommand = &cobra.Command{
	Use:   "g {rule_name}",
	Short: "Generate new rule",
	Args:  cobra.ExactArgs(1),
	Run: handleGenerate,
}
