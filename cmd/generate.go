package cmd

import (
	"errors"
	"fmt"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

const exampleRuleContents = `function main()
  return true
end
`

const rulesDirectory = "rules"

func handleGenerate(cmd *cobra.Command, args []string) {
	rulePath := strings.ReplaceAll(args[0], " ", "")
	if strings.HasSuffix(rulePath, ".lua") {
		rulePath = strings.TrimSuffix(rulePath, ".lua")
	}

	if rulePath == "" {
		printErrorAndExit(errors.New("the rule path cannot be empty"))
	}

	configFileContents, err := os.ReadFile(shared.ConfigFilename)
	if os.IsNotExist(err) {
		printErrorAndExit(errors.New("please run this command inside an actual spito ruleset directory"))
	}
	handleError(err)

	config := shared.ConfigFileLayout{}
	err = yaml.Unmarshal(configFileContents, &config)
	handleError(err)

	err = os.Mkdir(rulesDirectory, shared.DirectoryPermissions)
	if err != nil && !os.IsExist(err) {
		printErrorAndExit(err)
	}

	err = os.MkdirAll(filepath.Dir(rulePath), shared.DirectoryPermissions)
	handleError(err)

	ruleFile, err := os.Create(filepath.Join(rulesDirectory, rulePath+".lua"))
	handleError(err)
	_, err = ruleFile.WriteString(exampleRuleContents)
	handleError(err)
	err = ruleFile.Close()
	handleError(err)

	config.Rules[filepath.Base(rulePath)] = shared.RuleConfigLayout{
		Path:        filepath.Join(rulesDirectory, rulePath+".lua"),
		Description: "",
	}

	configFile, err := os.Create(shared.ConfigFilename)
	defer func() {
		handleError(configFile.Close())
	}()
	handleError(err)
	configFileContents, err = yaml.Marshal(config)
	handleError(err)
	_, err = configFile.Write(configFileContents)
	handleError(err)

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
