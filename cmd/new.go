package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/nasz-elektryk/spito/checker"
	"github.com/nasz-elektryk/spito/cmd/cmdApi"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var newRulesetCommand = &cobra.Command{
	Use:   "new {ruleset_name}",
	Short: "Create new ruleset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rulesetName := strings.ReplaceAll(args[0], " ", "")

		if rulesetName == "" {
			printErrorAndExit(errors.New("The ruleset name cannot be empty!"))
		}

		gitConfig, err := config.LoadConfig(config.GlobalScope)
		handleError(err)

		gitUsername := gitConfig.User.Name;
		if gitUsername == "" {
			printErrorAndExit(errors.New("Cannot find your git username. Please set it globally using git config"))
		}
		
		newRulesetLocation := checker.RuleSetLocation{}
		rulesetIdentifier := gitUsername + "/" + rulesetName
		newRulesetLocation.New(rulesetIdentifier)
		
		if newRulesetLocation.IsRuleSetDownloaded() {
			printErrorAndExit(errors.New(fmt.Sprintf("Ruleset %s already exists!", rulesetIdentifier)))
		}
		
		err = newRulesetLocation.CreateDir()
		handleError(err)

		_, err = git.PlainInit(newRulesetLocation.GetRuleSetPath(), false)
		handleError(err)

		filesToBeCreated := []string{"README.md", checker.LOCK_FILENAME}
		for _, fileName := range filesToBeCreated {
			file, err := os.Create(newRulesetLocation.GetRuleSetPath() + "/" + fileName)
			handleError(err)
			file.Close()
		}

		configFile, err := os.Create(newRulesetLocation.GetRuleSetPath() + "/" + checker.CONFIG_FILENAME)
		handleError(err)
		config := ConfigFileLayout{
			Repo_url: newRulesetLocation.GetFullUrl(),
			Git_prefix: checker.GetDefaultRepoPrefix(),
			Identifier: rulesetIdentifier,
			Rules: map[string]string{},
		}
		configFileContents, err := yaml.Marshal(config)
		handleError(err)
		configFile.Write(configFileContents)
		configFile.Close()

		err = os.Mkdir(newRulesetLocation.GetRuleSetPath() + "/" + "rules", 0700)
		handleError(err)

		infoApi := cmdApi.InfoApi{}
		infoApi.Log(fmt.Sprintf("Successfully created new ruleset '%s'", rulesetName))
	},
}
