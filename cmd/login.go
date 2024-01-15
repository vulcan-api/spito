package cmd

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/shared"
	"github.com/spf13/cobra"
)

const (
	tokenVerificationRoute = "token/verify"
	secretDirectoryName    = "secret"
	secretDirectoryPath    = "~/.local/state/spito/" + secretDirectoryName
	tokenStorageFilename   = "user-token"
)

var loginCommand = &cobra.Command{
	Use:   "login [-l|--local] {token}",
	Short: "Login into the spito store",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isLoggingInLocally, err := cmd.Flags().GetBool("local")
		handleError(err)

		if args[0] == "" {
			printErrorAndExit(errors.New("the token cannot be empty"))
		}

		if areWeInSpitoRuleset, _ := shared.DoesPathExist(checker.ConfigFilename); !areWeInSpitoRuleset && isLoggingInLocally {
			printErrorAndExit(errors.New("you must be inside a spito ruleset to log in locally"))
		}

		workingDirectory, err := os.Getwd()
		handleError(err)

		if exists, _ := shared.DoesPathExist(filepath.Join(workingDirectory, checker.ConfigFilename)); isLoggingInLocally && !exists {
			printErrorAndExit(errors.New("please run this command inside a spito ruleset"))
		}

		tokenRequestPath, err := url.JoinPath(os.Getenv("BACKEND_URL"), tokenVerificationRoute, args[0])
		handleError(err)

		httpResponse, err := http.Get(tokenRequestPath)
		handleError(err)
		defer httpResponse.Body.Close()

		var responseData map[string]interface{}
		json.NewDecoder(httpResponse.Body).Decode(&responseData)

		isTokenValid, isResponseOK := responseData["valid"]

		if !isResponseOK {
			httpResponse.Body.Close()
			printErrorAndExit(errors.New("the error has occured on the server side while validating the token"))
		}

		if !isTokenValid.(bool) && responseData["expiresAt"] != nil {
			httpResponse.Body.Close()
			cmdApi.InfoApi{}.Warn("Your token has expired. Please generate a new token and run this command again")
			os.Exit(1)
		} else if !isTokenValid.(bool) {
			httpResponse.Body.Close()
			printErrorAndExit(errors.New("your token is invalid. Please check if the token really belongs to your account"))
		}

		var secretDirectory string
		if isLoggingInLocally {
			secretDirectory = secretDirectoryName
		} else {
			secretDirectory = secretDirectoryPath
			shared.ExpandTilde(&secretDirectory)
		}

		err = os.MkdirAll(secretDirectory, 0755)
		if err != nil && !os.IsExist(err) {
			httpResponse.Body.Close()
			printErrorAndExit(errors.New("cannot create the directory for secrets"))
		}

		err = os.WriteFile(filepath.Join(secretDirectory, tokenStorageFilename), []byte(args[0]), 0644)
		handleError(err)

		cmdApi.InfoApi{}.Log("Successfully logged into the spito store!")
	},
}
