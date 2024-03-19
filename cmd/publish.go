package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/path"
	"github.com/avorty/spito/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	publishRulesetRoute  = "ruleset/publish"
	httpNotFoundCode     = 404
	httpOkLowerBoundCode = 200
	httpOkUpperBoundCode = 299
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

func isDescriptionDecorator(decoratorCode checker.RawDecorator) bool {
	return decoratorCode.Type == checker.DescriptionDecorator
}

func publishPostRequest(body PublishRequestBody, token string) int {
	jsonBody, err := json.Marshal(body)
	handleError(err)

	requestUrl, err := url.JoinPath(os.Getenv("BACKEND_URL"), publishRulesetRoute)
	handleError(err)

	httpRequest, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(jsonBody))
	handleError(err)

	httpRequest.Header.Add("Authorization", "Token "+token)
	httpRequest.Header.Add("Content-Type", "application/json")

	httpClient := http.Client{}

	httpResponse, err := httpClient.Do(httpRequest)
	handleError(err)
	defer func() {
		err = httpResponse.Body.Close()
		handleError(err)
	}()

	return httpResponse.StatusCode
}

func getToken(isLocal bool, rulesetPath string) string {
	tokenFilenamePath := filepath.Join(
		path.GetEnvWithDefaultValue("XDG_STATE_HOME", shared.LocalStateSpitoPath),
		secretDirectoryName,
		tokenStorageFilename)
	err := path.ExpandTilde(&tokenFilenamePath)
	handleError(err)

	tokenFileRawData, err := os.ReadFile(tokenFilenamePath)
	handleError(err)

	tokenData := TokenStorageLayout{}
	err = bson.Unmarshal(tokenFileRawData, &tokenData)
	handleError(err)

	var token string
	if isLocal {
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
	return token
}

func getDescriptionFromDecorator(script string, rulesetPath string, rulePath string) string {
	_, decorators, err := checker.GetDecorators(script)
	if err != nil {
		return ""
	}

	descriptionDecoratorPosition := slices.IndexFunc(decorators, isDescriptionDecorator)

	if descriptionDecoratorPosition == -1 {
		return ""
	}

	decorator := decorators[descriptionDecoratorPosition]
	description, pathArgument, err := checker.GetDecoratorArguments(decorator.Content)

	var result string
	path, ok := pathArgument["path"]
	if ok {
		descriptionBytes, err := os.ReadFile(filepath.Join(rulesetPath, rulePath, "..", path))
		handleError(err)
		result = string(descriptionBytes)
	} else if !ok && len(pathArgument) > 0 {
		rulePath := filepath.Join(rulesetPath, rulePath)
		err = errors.New("there's no 'path' argument specified inside Description decorator in rule: " + rulePath)
		printErrorAndExit(err)
	} else if len(description) > 0 {
		result = description[0]
	} else {
		printErrorAndExit(errors.New("incorrect 'Description' decorator syntax"))
	}
	return result
}

func onPublishCommand(cmd *cobra.Command, args []string) {
	var rulesetPath string
	var err error

	if len(args) > 0 && strings.TrimSpace(args[0]) != "" {
		rulesetPath = args[0]
	} else {
		rulesetPath, err = os.Getwd()
		handleError(err)
	}

	var configFileValues shared.ConfigFileLayout
	readConfigFile(rulesetPath, &configFileValues)

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

		descriptionFromDecorator := getDescriptionFromDecorator(string(ruleScript), rulesetPath, rule.Path)
		if descriptionFromDecorator != "" {
			currentRuleForRequest.Description = descriptionFromDecorator
		}
		requestBody.Rules = append(requestBody.Rules, currentRuleForRequest)
	}

	isTokenStoredLocally, err := cmd.Flags().GetBool("local")
	handleError(err)

	var token string
	if isTokenStoredLocally {
		token = getToken(true, rulesetPath)
	} else {
		token = getToken(false, "")
	}

	statusCode := publishPostRequest(requestBody, token)
	if statusCode == httpNotFoundCode {
		printErrorAndExit(errors.New("it seems that your ruleset doesn't exist in the spito store. Please create it in the spito GUI first"))
	} else if statusCode < httpOkLowerBoundCode || statusCode > httpOkUpperBoundCode {
		printErrorAndExit(errors.New("error during publishing"))
	}

	cmdApi.InfoApi{}.Log("successfully published the ruleset!")
}

var publishCommand = &cobra.Command{
	Use:   "publish [-l|--local] [ruleset_path]",
	Short: "Publish a ruleset to spito store",
	Args:  cobra.MaximumNArgs(1),
	Run:   onPublishCommand,
}
