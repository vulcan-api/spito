package cmd

import (
	"errors"
	"fmt"
	"github.com/avorty/spito/internal/checker"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func getGitUsername() string {
	gitConfig, err := config.LoadConfig(config.GlobalScope)
	handleError(err)
	gitUsername := gitConfig.User.Name
	if gitUsername == "" {
		printErrorAndExit(errors.New("Cannot find your git username. Please set it globally using git config"))
	}
	return gitUsername
}

func isRequestPathOK(urlToValidate url.URL) bool {
	result, err := regexp.Match("/[a-zA-z0-9\\-_]+/[a-zA-z0-9\\-_]+/?", []byte(urlToValidate.Path))
	handleError(err)
	return result
}

var newRulesetCommand = &cobra.Command{
	Use:   "new [-y] {ruleset_name}",
	Short: "Create new ruleset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rulesetName := strings.ReplaceAll(args[0], " ", "")

		if rulesetName == "" {
			printErrorAndExit(errors.New("The ruleset name cannot be empty!"))
		}

		_, err := os.Stat(rulesetName)
		if err == nil {
			printErrorAndExit(errors.New(fmt.Sprintf("Ruleset '%s' already exists!", rulesetName)))
		}

		shouldAssumeDefaultValues, err := cmd.Flags().GetBool("non-interactive")
		handleError(err)

		gitUsername := getGitUsername()
		rulesetIdentifier := gitUsername + "/" + rulesetName
		newRulesetLocation := checker.NewRulesetLocation(rulesetIdentifier)

		// Because we create RulesetLocation based on git identifier, we can be sure that full url is not nil
		repositoryUrl := *newRulesetLocation.GetFullUrl()
		rulesetRepositoryName := rulesetName
		hostingProvider := checker.GetDefaultRepoPrefix()

		if !shouldAssumeDefaultValues {
			var input string

			fmt.Printf("Enter your git service username (%s): ", gitUsername)
			fmt.Scanf("%s", &input)
			if input != "" {
				gitUsername = input
			}
			input = ""

			fmt.Printf("Enter your ruleset repository name (%s): ", rulesetName)
			fmt.Scanf("%s", &input)
			if input != "" {
				rulesetRepositoryName = input
			}
			input = ""

			fmt.Printf("Enter your git repository hosting provider (%s): ", checker.GetDefaultRepoPrefix())
			fmt.Scanf("%s", &input)
			if input != "" {
				hostingProvider = input
			}
			input = ""

			repositoryUrl = fmt.Sprintf("https://%s/%s/%s", hostingProvider, gitUsername, rulesetRepositoryName)
			fmt.Printf("Enter repository URL (%s): ", repositoryUrl)
			fmt.Scanf("%s", &input)
			if input != "" {
				repositoryUrlObject, err := url.ParseRequestURI(input)
				for err != nil || !isRequestPathOK(*repositoryUrlObject) {
					fmt.Print("Enter a valid URL: ")
					fmt.Scanf("%s", &input)
					repositoryUrlObject, err = url.ParseRequestURI(input)
				}
				if input[len(repositoryUrl)-1] == '/' {
					input = strings.TrimRight(repositoryUrl, "/")
					repositoryUrlObject.Path = strings.TrimRight(repositoryUrlObject.Path, "/")
				}
				repositoryUrl = input
			}
			rulesetIdentifier = gitUsername + "/" + rulesetRepositoryName
		}

		err = os.Mkdir(rulesetName, 0700)
		handleError(err)

		_, err = git.PlainInit(rulesetName, false)
		handleError(err)

		filesToBeCreated := []string{"README.md", checker.LockFilename}
		for _, fileName := range filesToBeCreated {
			file, err := os.Create(rulesetName + "/" + fileName)
			handleError(err)
			file.Close()
		}

		configFile, err := os.Create(rulesetName + "/" + checker.ConfigFilename)
		handleError(err)
		config := ConfigFileLayout{
			Repo_url:   repositoryUrl,
			Git_prefix: hostingProvider,
			Identifier: rulesetIdentifier,
			Rules:      map[string]string{},
		}
		configFileContents, err := yaml.Marshal(config)
		handleError(err)
		configFile.Write(configFileContents)
		configFile.Close()

		err = os.Mkdir(rulesetName+"/"+"rules", 0700)
		handleError(err)

		infoApi := cmdApi.InfoApi{}
		infoApi.Log(fmt.Sprintf("Successfully created new ruleset '%s'", rulesetName))
	},
}
