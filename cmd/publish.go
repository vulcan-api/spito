package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

const (
	publishRulesetRoute = "ruleset/publish"
)

type RuleForRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Unsafe      bool   `json:"unsafe"`
}

type PublishRequestBody struct {
	Url    string           `json:"url"`
	Branch string           `json:"branch"`
	Rules  []RuleForRequest `json:"rules"`
}

var publishCommand = &cobra.Command{
	Use:   "publish [-l|--local] [ruleset_path]",
	Short: "Publish a ruleset to spito store",
	Args:  cobra.MaximumNArgs(1),
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
			printErrorAndExit(errors.New("please point this command to an actual spito ruleset"))
		} else {
			handleError(err)
		}

		var configFileValues ConfigFileLayout
		err = yaml.Unmarshal(configFileContents, &configFileValues)
		handleError(err)

		if configFileValues.RepoUrl == "" || configFileValues.Branch == "" {
			printErrorAndExit(errors.New("repo_url and branch keys in spito configuration file cannot be empty"))
		}

		requestBody := PublishRequestBody{
			Url:    configFileValues.RepoUrl,
			Branch: configFileValues.Branch,
		}
		for ruleName, rule := range configFileValues.Rules {
			currentRuleForRequest := RuleForRequest{
				Name:        ruleName,
				Path:        rule.Path,
				Description: rule.Description,
				Unsafe:      rule.Unsafe,
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

				description, pathArgument, err := checker.GetDecoratorArguments(decorator)

				if len(pathArgument) > 0 {
					path, ok := pathArgument["path"]
					if !ok {
						printErrorAndExit(
							errors.New("there's no 'path' argument specified inside Description decorator in rule: " +
								filepath.Join(rulesetPath, rule.Path)))
					}
					descriptionBytes, err := os.ReadFile(filepath.Join(rulesetPath, rule.Path, "..", path))
					handleError(err)
					currentRuleForRequest.Description = string(descriptionBytes)
					break
				} else if len(description) > 0 {
					currentRuleForRequest.Description = description[0]
				} else {
					printErrorAndExit(errors.New("incorrect 'Description' decorator syntax"))
				}
				break
			}

			requestBody.Rules = append(requestBody.Rules, currentRuleForRequest)
		}

		var tokenFilenamePath string

		tokenFilenamePath = filepath.Join(
			shared.GetEnvWithDefaultValue("XDG_STATE_HOME", secretGlobalDirectoryDefaultValue),
			secretDirectoryName,
			tokenStorageFilename)
		err = shared.ExpandTilde(&tokenFilenamePath)
		handleError(err)

		isTokenStoredLocally, err := cmd.Flags().GetBool("local")
		handleError(err)

		tokenFileRawData, err := os.ReadFile(tokenFilenamePath)
		handleError(err)

		tokenData := TokenStorageLayout{}
		err = bson.Unmarshal(tokenFileRawData, &tokenData)
		handleError(err)

		var token string
		if isTokenStoredLocally {
			tokenPosition := slices.IndexFunc(tokenData.LocalKeys, func(localToken LocalToken) bool {
				return localToken.Path == rulesetPath
			})
			if tokenPosition == -1 {
				printErrorAndExit(errors.New("cannot find the local token for ruleset: " + rulesetPath))
			}
			token = tokenData.LocalKeys[tokenPosition].Token
		} else {
			token = tokenData.GlobalToken
		}

		jsonBody, err := json.Marshal(requestBody)
		handleError(err)

		requestUrl, err := url.JoinPath(os.Getenv("BACKEND_URL"), publishRulesetRoute)
		handleError(err)

		httpRequest, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonBody))
		handleError(err)

		httpRequest.Header.Add("Authorization", "Token "+string(token))
		httpRequest.Header.Add("Content-Type", "application/json")

		httpClient := http.Client{}

		httpResponse, err := httpClient.Do(httpRequest)
		handleError(err)
		defer func() {
			err = httpResponse.Body.Close()
			handleError(err)
		}()

		var responseBody map[string]interface{}
		err = json.NewDecoder(httpResponse.Body).Decode(&responseBody)

		if statusCode, codeExists := responseBody["statusCode"]; codeExists && statusCode.(float64) == 404 {
			printErrorAndExit(errors.New("it seems that your ruleset doesn't exist in the spito store. Please create it in the spito GUI first"))
		} else if codeExists && statusCode.(float64) != 200 {
			printErrorAndExit(errors.New("error during publishing"))
		}

		cmdApi.InfoApi{}.Log("successfully published the ruleset!")
	},
}
