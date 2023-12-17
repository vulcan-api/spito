package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"github.com/go-git/go-git/v5"
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
		gitCommand := exec.Command("git", "config", "--global", "--list")

		gitCommandOutput, error := gitCommand.Output()
		if error != nil {
			printErrorAndExit(errors.New("Cannot find git command in your PATH. Please install git"))
		}

		lines := strings.Split(string(gitCommandOutput), "\n")
		var gitUsername string
		for _, line := range lines {
			tokens := strings.Split(line, "=")
			if tokens[0] == "user.name" {
				gitUsername = tokens[1]
				break
			}
		}

		if gitUsername == "" {
			printErrorAndExit(errors.New("Cannot find your git username. Please set your username globally"))
		}
		
		newRulesetLocation := checker.RuleSetLocation{}
		rulesetIdentifier := gitUsername + "/" + rulesetName
		newRulesetLocation.New(rulesetIdentifier)
		
		if newRulesetLocation.IsRuleSetDownloaded() {
			printErrorAndExit(errors.New(fmt.Sprintf("Ruleset %s already exists!", rulesetIdentifier)))
		}
		
		error = newRulesetLocation.CreateDir()
		handleError(error)

		_, error = git.PlainInit(newRulesetLocation.GetRuleSetPath(), false)
		handleError(error)

		filesToBeCreated := []string{"README.md", checker.LOCK_FILENAME}
		for _, fileName := range filesToBeCreated {
			file, error := os.Create(newRulesetLocation.GetRuleSetPath() + "/" + fileName)
			handleError(error)
			file.Close()
		}

		configFile, error := os.Create(newRulesetLocation.GetRuleSetPath() + "/" + checker.CONFIG_FILENAME)
		handleError(error)
		config := ConfigFileLayout{
			Repo_url: newRulesetLocation.GetFullUrl(),
			Git_prefix: checker.GetDefaultRepoPrefix(),
			Identifier: rulesetIdentifier,
			Rules: map[string]string{},
		}
		configFileContents, error := yaml.Marshal(config)
		handleError(error)
		configFile.Write(configFileContents)
		configFile.Close()

		error = os.Mkdir(newRulesetLocation.GetRuleSetPath() + "/" + "rules", os.ModeDir)
		handleError(error)

		infoApi := cmdApi.InfoApi{}
		infoApi.Log(fmt.Sprintf("Successfully created new ruleset '%s'", rulesetName))
	},
}
