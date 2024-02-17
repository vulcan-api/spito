package cmd

import (
	"fmt"
	"github.com/avorty/spito/internal/checker"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var envFileCmd = &cobra.Command{
	Use:   "file {path}",
	Short: "Applies specified environment",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runtimeData := getInitialRuntimeData(cmd)
		envScriptPath := args[0]

		defer func() {
			if err := runtimeData.DeleteRuntimeTemp(); err != nil {
				fmt.Printf("Failed to remove temporary VRCT files"+
					"\n You should remove them manually in /tmp or reboot your device \n%s", err.Error())
				os.Exit(1)
			}
		}()

		envScriptPathAbs, err := filepath.Abs(envScriptPath)
		if err != nil {
			fmt.Printf("Failed to convert %s to absolute path\n", envScriptPath)
			fmt.Println(err.Error())
			os.Exit(1)
		}

		envScript, err := os.ReadFile(envScriptPathAbs)
		handleError(err)

		err = checker.ApplyEnvironmentScript(&runtimeData, string(envScript), envScriptPathAbs)
		handleError(err)

		fmt.Printf("Successfully applied %s environment\n", envScriptPath)
	},
}

var envCmd = &cobra.Command{
	Use:   "env {ruleset identifier or path} {env}",
	Short: "Applies specified environment",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runtimeData := getInitialRuntimeData(cmd)
		identifierOrPath := args[0]
		envName := args[1]

		defer func() {
			if err := runtimeData.DeleteRuntimeTemp(); err != nil {
				fmt.Printf("Failed to remove temporary VRCT files"+
					"\n You should remove them manually in /tmp or reboot your device \n%s", err.Error())
				os.Exit(1)
			}
		}()

		err := checker.ApplyEnvironmentByIdentifier(&runtimeData, identifierOrPath, envName)
		handleError(err)

		fmt.Printf("Successfully applied %s environment\n", envName)
	},
}
