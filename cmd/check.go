package cmd

import (
	"bufio"
	"fmt"
	"github.com/avorty/spito/pkg/api"
	"github.com/avorty/spito/pkg/package_conflict"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/cmd/guiApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/godbus/dbus"
	"github.com/spf13/cobra"
)

func askAndExecuteRule(runtimeData shared.ImportLoopData) {
	fmt.Printf("Would you like to execute this rule? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	answer, _, err := reader.ReadRune()
	handleError(err)

	answer = unicode.ToLower(answer)

	if answer != 'y' {
		return
	}

	revertNum, err := runtimeData.VRCT.Apply()
	if err != nil {
		err = runtimeData.VRCT.Revert()
		runtimeData.InfoApi.Error("unfortunately the rule couldn't be applied. Reverting changes...")
		handleError(err)
	}

	packagesToInstall := runtimeData.PackageTracker.GetPackagesToInstall()
	packagesToRemove := runtimeData.PackageTracker.GetPackagesToRemove()

	if len(packagesToInstall) > 0 {
		runtimeData.InfoApi.Log("installing packages...")
		err = api.InstallPackage(strings.Join(packagesToInstall, " "))
		handleError(err)
	}

	if len(packagesToRemove) > 0 {
		runtimeData.InfoApi.Log("removing packages...")
		err = api.RemovePackage(strings.Join(packagesToRemove, " "))
		handleError(err)
	}

	revertCommand := fmt.Sprintf("spito revert %d", revertNum)
	runtimeData.InfoApi.Log("In order to revert changes, use this command: ", revertCommand)
}

var checkFileCmd = &cobra.Command{
	Use:   "file {path}",
	Short: "Check local lua rule file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		runtimeData := getInitialRuntimeData(cmd)
		defer func() {
			if err := runtimeData.DeleteRuntimeTemp(); err != nil {
				fmt.Printf("Failed to remove temporary VRCT files"+
					"\n You should remove them manually in /tmp or reboot your device \n%s", err.Error())
				os.Exit(1)
			}
		}()

		script, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Failed to read file %s\n", path)
			os.Exit(1)
		}

		fileAbsolutePath, err := filepath.Abs(path)
		if err != nil {
			runtimeData.InfoApi.Error("Cannot create the absolute path to the file!")
			os.Exit(1)
		}

		doesRulePass, err := checker.CheckRuleScript(&runtimeData, string(script), filepath.Dir(fileAbsolutePath))
		if err != nil {
			panic(err)
		}

		communicateRuleResult(path, doesRulePass)

		if doesRulePass {
			askAndExecuteRule(runtimeData)
		}
	},
}

var checkCmd = &cobra.Command{
	Use:   "check {ruleset identifier or path} {rule}",
	Short: "Check whether your machine pass rule",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		runtimeData := getInitialRuntimeData(cmd)

		identifierOrPath := args[0]
		ruleName := args[1]

		isPath, err := shared.DoesPathExist(identifierOrPath)
		handleError(err)

		defer func() {
			if err := runtimeData.DeleteRuntimeTemp(); err != nil {
				fmt.Printf("Failed to remove temporary VRCT files"+
					"\n You should remove them manually in /tmp or reboot your device \n%s", err.Error())
				os.Exit(1)
			}
		}()

		var doesRulePass bool
		if isPath {
			doesRulePass, err = checker.CheckRuleByPath(&runtimeData, identifierOrPath, ruleName)
		} else {
			doesRulePass, err = checker.CheckRuleByIdentifier(&runtimeData, identifierOrPath, ruleName)
		}
		handleError(err)

		communicateRuleResult(ruleName, doesRulePass)
		if doesRulePass {
			askAndExecuteRule(runtimeData)
		}
	},
}

func getInitialRuntimeData(cmd *cobra.Command) shared.ImportLoopData {
	isExecutedByGui, err := cmd.Flags().GetBool("gui-child-mode")
	if err != nil {
		isExecutedByGui = true
	}

	options, err := cmd.Flags().GetStringArray("options")
	if err != nil {
		options = nil
	}

	var infoApi shared.InfoInterface

	if isExecutedByGui {
		conn, err := dbus.SessionBus()
		if err != nil {
			panic(err)
		}

		dbusId := os.Getenv("DBUS_INTERFACE_ID")
		dbusPath := os.Getenv("DBUS_OBJECT_PATH")

		busObject := conn.Object(dbusId, dbus.ObjectPath(dbusPath))
		infoApi = guiApi.InfoApi{
			BusObject: busObject,
		}
	} else {
		infoApi = cmdApi.InfoApi{}
	}

	ruleVRCT, err := vrct.NewRuleVRCT()
	if err != nil {
		panic(err)
	}

	return shared.ImportLoopData{
		VRCT:           *ruleVRCT,
		RulesHistory:   shared.RulesHistory{},
		ErrChan:        make(chan error),
		InfoApi:        infoApi,
		PackageTracker: package_conflict.NewPackageConflictTracker(),
		Options:        options,
	}
}

func communicateRuleResult(ruleName string, doesRulePass bool) {
	if doesRulePass {
		fmt.Printf("Rule %s successfuly passed requirements\n", ruleName)
	} else {
		fmt.Printf("Rule %s did not pass requirements\n", ruleName)
	}
}
