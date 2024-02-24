package cmd

import (
	"encoding/json"
	"errors"
	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/pkg/shared"
	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2/bson"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	tokenVerificationRoute = "token/verify"
	secretDirectoryName    = "secret"
	tokenStorageFilename   = "user-tokens.json"
)

type TokenValidationResponse struct {
	Valid     bool    `json:"valid"`
	Message   string  `json:"message"`
	ExpiresAt *string `json:"expiresAt"`
}

type LocalToken struct {
	Token string `json:"token"`
	Path  string `json:"path"`
}

type TokenStorageLayout struct {
	GlobalToken string       `json:"globalToken"`
	LocalKeys   []LocalToken `json:"localKeys"`
}

func onLoginCommand(cmd *cobra.Command, args []string) {
	isLoggingInLocally, err := cmd.Flags().GetBool("local")
	handleError(err)

	if strings.TrimSpace(args[0]) == "" {
		printErrorAndExit(errors.New("the token cannot be empty"))
	}

	if exists, _ := shared.PathExists(shared.ConfigFilename); isLoggingInLocally && !exists {
		printErrorAndExit(errors.New("please run this command inside a spito ruleset"))
	}

	tokenRequestPath, err := url.JoinPath(os.Getenv("BACKEND_URL"), tokenVerificationRoute, args[0])
	handleError(err)

	httpResponse, err := http.Get(tokenRequestPath)
	handleError(err)
	defer func() {
		err = httpResponse.Body.Close()
		handleError(err)
	}()

	responseData := TokenValidationResponse{}
	err = json.NewDecoder(httpResponse.Body).Decode(&responseData)
	handleError(err)

	if !responseData.Valid && responseData.ExpiresAt != nil {
		err = httpResponse.Body.Close()
		handleError(err)
		cmdApi.InfoApi{}.Warn("your token has expired. Please generate a new token and run this command again")
		os.Exit(1)

	} else if !responseData.Valid {
		err = httpResponse.Body.Close()
		handleError(err)

		printErrorAndExit(errors.New("your token is invalid. Please check if the token really belongs to your account"))
	}
	secretFilePath := filepath.Join(
		shared.GetEnvWithDefaultValue("XDG_STATE_HOME", shared.LocalStateSpitoPath),
		secretDirectoryName,
		tokenStorageFilename)
	err = shared.ExpandTilde(&secretFilePath)
	handleError(err)
	err = os.MkdirAll(filepath.Dir(secretFilePath), shared.DirectoryPermissions)

	if err != nil && !os.IsExist(err) {
		err = httpResponse.Body.Close()
		handleError(err)
		printErrorAndExit(errors.New("cannot create the directory for secrets"))
	}

	tokenData := TokenStorageLayout{}

	if doesTokenFileExists, _ := shared.PathExists(secretFilePath); doesTokenFileExists {
		tokenFileRaw, err := os.ReadFile(secretFilePath)
		handleError(err)
		err = bson.Unmarshal(tokenFileRaw, &tokenData)
		handleError(err)
	}

	if isLoggingInLocally {
		workingDirectory, err := os.Getwd()
		handleError(err)

		tokenData.LocalKeys = append(tokenData.LocalKeys, LocalToken{
			Token: args[0],
			Path:  workingDirectory,
		})
	} else {
		tokenData.GlobalToken = args[0]
	}

	bsonOutput, err := bson.Marshal(tokenData)
	handleError(err)

	err = os.WriteFile(secretFilePath, bsonOutput, shared.FilePermissions)
	handleError(err)

	cmdApi.InfoApi{}.Log("successfully logged into the spito store")
}

var loginCommand = &cobra.Command{
	Use:   "login [-l|--local] {token}",
	Short: "Login into the spito store",
	Args:  cobra.ExactArgs(1),
	Run:   onLoginCommand,
}
