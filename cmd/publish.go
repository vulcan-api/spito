package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	publishRulesetRoute = "ruleset/publish"
)

type RuleForRequest struct {
	Name string `json:"name"`
	Description string `json:"description"`
	Path string `json:"path"`
}

type PublishRequestBody struct {
	Url string `json:"url"`
	Rules []RuleForRequest `json:"rules"`
}

var publishCommand = &cobra.Command{
	Use:   "publish [-l|--local] [ruleset_path]",
	Short: "Publish a ruleset to spito store",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var rulesetPath string
		var err error
		
		if len(args) > 0 && args[0] != "" {
			rulesetPath = args[0]
		} else {
			rulesetPath, err = os.Getwd()
			handleError(err)
		}

		configFileContents, err := os.ReadFile(filepath.Join(rulesetPath, checker.ConfigFilename))
		if os.IsNotExist(err) {
			printErrorAndExit(errors.New("Please point this command to an actual spito ruleset!"))
		} else {
			handleError(err)
		}

		var configFileValues ConfigFileLayout
		err = yaml.Unmarshal(configFileContents, &configFileValues)
		handleError(err)

		requestBody := PublishRequestBody {
			Url:  configFileValues.Repo_url,
		}
		for ruleName, rule := range configFileValues.Rules {
			currentRuleForRequest := RuleForRequest{
				Name: ruleName,
				Path: rule.Path,
				Description: rule.Description,
			}

			ruleScript, err := os.ReadFile(filepath.Join(rulesetPath, rule.Path))
			handleError(err)

			_, decorators := checker.GetDecorators(string(ruleScript))
			for _, decorator := range decorators {
				isDescriptionDecorator, err := regexp.Match(`^Description\(.*\)$`, []byte(decorator))
				handleError(err)

				if !isDescriptionDecorator {
					continue
				}

				betweenParanthesesRegex := regexp.MustCompile(`\(.*\)`)
				argument := betweenParanthesesRegex.FindString(decorator)
				argument = strings.TrimSpace(argument)
				argument = strings.TrimPrefix(argument, "(")
				argument = strings.TrimSuffix(argument, ")")

				if strings.HasPrefix(argument, "file=") {
					tokens := strings.Split(argument, "=")
					if len(tokens) < 2 {
						printErrorAndExit(errors.New("Incorrect \"Description\" decorator syntax inside rule: " + ruleName))
					}
					descriptionBytes, err := os.ReadFile(filepath.Join(rule.Path,"..",tokens[1][1:len(tokens[1]) - 1]))
					handleError(err)

					currentRuleForRequest.Description = string(descriptionBytes)
				} else if strings.HasSuffix(argument, "\"") && strings.HasPrefix(argument, "\"") {
					currentRuleForRequest.Description = argument[1:len(argument) - 1]
				} else {
					printErrorAndExit(errors.New("Incorrect \"Description\" decorator syntax inside rule: " + ruleName))
				}
				break
			}
			
			requestBody.Rules = append(requestBody.Rules, currentRuleForRequest)
		}

		var tokenFilenamePath string
		isTokenStoredLocally, err := cmd.Flags().GetBool("local")
		handleError(err)
		
		if isTokenStoredLocally {
			tokenFilenamePath = filepath.Join(rulesetPath, secretDirectoryName, tokenStorageFilename)
		} else {
			tokenFilenamePath = filepath.Join(secretDirectoryPath, tokenStorageFilename)
			shared.ExpandTilde(&tokenFilenamePath)
		}

		token, err := os.ReadFile(tokenFilenamePath)
		handleError(err)

		jsonBody, err := json.Marshal(requestBody)
		handleError(err)

		requestUrl, err := url.JoinPath(spitoStoreURL, publishRulesetRoute)
		handleError(err)
		
		httpRequest, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonBody))
		handleError(err)

		httpRequest.Header.Add("Authorization", "Token " + string(token))
		httpRequest.Header.Add("Content-Type", "application/json")

		httpClient := http.Client{}

		httpResponse, err := httpClient.Do(httpRequest)
		handleError(err)
		defer httpResponse.Body.Close()

		var responseBody map[string]interface{}
		json.NewDecoder(httpResponse.Body).Decode(&responseBody)

		if statusCode, codeExists := responseBody["statusCode"]; codeExists && statusCode.(float64) == 404 {
			printErrorAndExit(errors.New("It seems that your ruleset doesn't exist in the spito store. Please create it in the spito GUI first"))
		} else if codeExists && statusCode.(float64) != 200 {
			printErrorAndExit(errors.New("Error during publishing!"))
		}

		cmdApi.InfoApi{}.Log("Successfully published the ruleset!")
	},
}
