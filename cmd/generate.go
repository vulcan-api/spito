package cmd

import (
	"errors"
	"fmt"
	"github.com/nasz-elektryk/spito/cmd/cmdApi"
	"github.com/nasz-elektryk/spito/internal/checker"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const exampleRuleContents = `function main()
  return true
end
`

const RULES_DIRECTORY = "rules"

func handleGenerate(cmd *cobra.Command, args []string) {
	rulePath := strings.ReplaceAll(args[0], " ", "")

	if rulePath == "" {
		printErrorAndExit(errors.New("The rule path cannot be empty!"))
	}

	configFileContents, err := os.ReadFile(checker.CONFIG_FILENAME)
	if os.IsNotExist(err) {
		printErrorAndExit(errors.New("Please run this commnd inside an actual spito ruleset directory!"))
	}
	handleError(err)

	config := ConfigFileLayout{}
	err = yaml.Unmarshal(configFileContents, &config)
	handleError(err)

	rulesetLocation := checker.RuleSetLocation{}
	rulesetLocation.New(config.Identifier)

	err = os.Mkdir(RULES_DIRECTORY, 0700)
	if err != nil && !os.IsExist(err) {
		printErrorAndExit(err)
	}

	rulePathTokens := strings.Split(rulePath, "/")
	if len(rulePathTokens) > 1 {
		directoryPath := RULES_DIRECTORY + "/" + strings.Join(rulePathTokens[:len(rulePathTokens)-1], "/")
		err = os.MkdirAll(directoryPath, 0700)
		handleError(err)
	}

	ruleFile, err := os.Create(RULES_DIRECTORY + "/" + rulePath + ".lua")
	handleError(err)
	ruleFile.Write([]byte(exampleRuleContents))
	ruleFile.Close()

	config.Rules[rulePathTokens[len(rulePathTokens)-1]] = "./rules/" + rulePath + ".lua"

	configFile, err := os.Create(checker.CONFIG_FILENAME)
	defer configFile.Close()
	handleError(err)
	configFileContents, err = yaml.Marshal(config)
	handleError(err)
	configFile.Write(configFileContents)

	infoApi := cmdApi.InfoApi{}
	infoApi.Log(fmt.Sprintf("Successfully created rule '%s'", rulePath))
}

var generateRuleCommand = &cobra.Command{
	Use:   "generate {rule_path}",
	Short: "Generate new rule",
	Args:  cobra.ExactArgs(1),
	Run:   handleGenerate,
}

var generateShortCommand = &cobra.Command{
	Use:   "g {rule_path}",
	Short: "Generate new rule",
	Args:  cobra.ExactArgs(1),
	Run:   handleGenerate,
}
