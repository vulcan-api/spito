package main

import (
	_ "embed"
	"fmt"
	"github.com/avorty/spito/cmd"
	"github.com/joho/godotenv"
	"github.com/avorty/spito/pkg/userinfo"
	"os"
)

func main() {
	userinfo.ChangeToUser()
	initEnvs()
	cmd.Execute()
}

//go:embed build.env
var buildEnvContent string

func initEnvs() {
	envs, err := godotenv.Unmarshal(buildEnvContent)
	if err != nil {
		fmt.Println("Failed to parse build.env \n", err.Error())
		os.Exit(1)
	}

	for key, value := range envs {
		err := os.Setenv(key, value)
		if err != nil {
			fmt.Printf("Failed to set following env variable: %s=%s \n%s", key, value, err.Error())
			os.Exit(1)
		}
	}
}
