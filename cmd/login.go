package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/shared"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	spitoStoreURL = "http://localhost:5000"
	loginRoute = "auth/login"
	tokenCreationRoute = "token"
	secretDirectoryName = ".secret"
	tokenStorageFilename = "user-token"
)

var loginCommand = &cobra.Command{
	Use:   "login",
	Short: "Login into the spito store",
	Run: func(cmd *cobra.Command, args []string) {

		workingDirectory, err := os.Getwd()
		handleError(err)
		
		if exists, _ := shared.DoesPathExist(filepath.Join(workingDirectory, checker.ConfigFilename)); !exists {
			printErrorAndExit(errors.New("Please run this command inside a spito ruleset!"))
		}
		
		var email string
		for {
			fmt.Printf("Enter your email address: ")
			fmt.Scanf("%s", &email)
			if email != "" {
				break
			}
			fmt.Println("Email address cannot be empty!")
		}

		var password []byte
		for {
			fmt.Printf("Enter your password: ")
			password, err = term.ReadPassword(int(os.Stdin.Fd()))
			handleError(err)
			if string(password) != "" {
				fmt.Println()
				break
			}
			fmt.Println("Password cannot be empty!")
		}
		loginRequestPath, err := url.JoinPath(spitoStoreURL, loginRoute)
		handleError(err)
		
		body := map[string]string {
			"email": email,
			"password": string(password),
		}

		jsonBody, err := json.Marshal(body)
		handleError(err)
		
		httpResponse, err := http.Post(loginRequestPath, "application/json", bytes.NewBuffer(jsonBody))
		handleError(err)

		var responseData map[string]string

		json.NewDecoder(httpResponse.Body).Decode(&responseData)
		if _, doesTokenExist := responseData["token"]; !doesTokenExist {
			httpResponse.Body.Close()
			printErrorAndExit(errors.New("Login has failed. Please provide correct credentials"))
		}
		httpResponse.Body.Close()

		tokenRequestPath, err := url.JoinPath(spitoStoreURL, tokenCreationRoute)
		handleError(err)
		
		expirationTime := time.Now()
		expirationTime = expirationTime.AddDate(0, 0, 7)
		
		body = map[string]string {
			"expiresAt": expirationTime.Format(time.RFC3339),
		}
		jsonBody, err = json.Marshal(body)
		handleError(err)
		
		httpRequest, err := http.NewRequest("POST", tokenRequestPath, bytes.NewReader(jsonBody))
		handleError(err)
		httpRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", responseData["token"]))
		httpRequest.Header.Add("Content-Type", "application/json")
		
		httpClient := http.Client{}
		httpResponse, err = httpClient.Do(httpRequest)
		handleError(err)
		defer httpResponse.Body.Close()
		
		responseData = make(map[string]string)
		json.NewDecoder(httpResponse.Body).Decode(&responseData)
		if _, isResponseOK := responseData["token"]; !isResponseOK {
			httpResponse.Body.Close()
			printErrorAndExit(errors.New("Couldn't generate the token"))
		}

		err = os.Mkdir(secretDirectoryName, 0700)
		if err != nil && !os.IsExist(err) {
			httpResponse.Body.Close()
			printErrorAndExit(errors.New("Cannot create the directory for secrets!"))
		}

		err = os.WriteFile(filepath.Join(secretDirectoryName, tokenStorageFilename), []byte(responseData["token"]), 0644)
		handleError(err)

		cmdApi.InfoApi{}.Log("Successfully logged into the spito store!")
	},
}
