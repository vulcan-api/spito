package cmd

import (
	"bytes"
	"fmt"
	"github.com/avorty/spito/pkg/package_conflict"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/avorty/spito/cmd/cmdApi"
	"github.com/avorty/spito/cmd/guiApi"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/shared"
	"github.com/avorty/spito/pkg/vrct"
	"github.com/godbus/dbus/v5"
	"github.com/spf13/cobra"
)

func askAndExecuteRule(runtimeData shared.ImportLoopData, guiMode bool) {
	revertNum, err := runtimeData.VRCT.Apply()
	if err != nil {
		err = runtimeData.VRCT.Revert()
		runtimeData.InfoApi.Error("unfortunately the rule couldn't be applied. Reverting changes...")
		handleError(err)
	}

	if guiMode {
		shared.DBusMethodP(runtimeData.DbusConn, "Success", "cannot send success message", revertNum)
	} else {
		revertCommand := fmt.Sprintf("spito revert %d", revertNum)
		runtimeData.InfoApi.Log("In order to revert changes, use this command:", revertCommand)
	}

}

var checkFileCmd = &cobra.Command{
	Use:   "file {path}",
	Short: "Check local lua rule file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputPath := args[0]

		runtimeData := getInitialRuntimeData(cmd)
		defer func() {
			if err := runtimeData.DeleteRuntimeTemp(); err != nil {
				fmt.Printf("Failed to remove temporary VRCT files"+
					"\n You should remove them manually in /tmp or reboot your device \n%s", err.Error())
				os.Exit(1)
			}
		}()

		script, err := os.ReadFile(inputPath)
		if err != nil {
			fmt.Printf("Failed to read file %s\n", inputPath)
			os.Exit(1)
		}

		fileAbsolutePath, err := filepath.Abs(inputPath)
		if err != nil {
			runtimeData.InfoApi.Error("Cannot create the absolute inputPath to the file!")
			os.Exit(1)
		}

		ruleConf, err := checker.GetRuleConfFromScript(fileAbsolutePath)
		handleError(err)
		panicIfEnvironment(runtimeData, &ruleConf, "file", inputPath)

		doesRulePass, err := checker.CheckRuleScript(&runtimeData, string(script), filepath.Dir(fileAbsolutePath))
		if err != nil {
			panic(err)
		}

		if doesRulePass {
			askAndExecuteRule(runtimeData, false)
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

		isPath, err := shared.PathExists(identifierOrPath)
		handleError(err)

		defer func() {
			if err := runtimeData.DeleteRuntimeTemp(); err != nil {
				fmt.Printf("Failed to remove temporary VRCT files"+
					"\n You should remove them manually in /tmp or reboot your device \n%s", err.Error())
				os.Exit(1)
			}
		}()

		rulesetLocation := checker.NewRulesetLocation(identifierOrPath, isPath)
		rulesetConfig, err := checker.GetRulesetConf(&rulesetLocation)
		handleError(err)

		ruleConf, err := rulesetConfig.GetRuleConf(ruleName)
		handleError(err)
		panicIfEnvironment(runtimeData, &ruleConf, identifierOrPath, ruleName)

		var doesRulePass bool
		if isPath {
			doesRulePass, err = checker.CheckRuleByPath(&runtimeData, identifierOrPath, ruleName)
		} else {
			doesRulePass, err = checker.CheckRuleByIdentifier(&runtimeData, identifierOrPath, ruleName)
		}
		handleError(err)

		if runtimeData.GuiMode {
			err := runtimeData.DbusConn.AddMatchSignal(
				dbus.WithMatchObjectPath(shared.DBusObjectPath()),
				dbus.WithMatchInterface(shared.DBusInterfaceId()),
				dbus.WithMatchSender(shared.DBusInterfaceId()))
			if err != nil {
				panic(err)
			}
			shared.DBusMethodP(runtimeData.DbusConn, "CheckFinished", "cannot connect to gui", doesRulePass)
			replyChan := make(chan *dbus.Signal)
			runtimeData.DbusConn.Signal(replyChan)
			reply := <-replyChan
			if reply.Name != shared.DBusInterfaceId()+".Confirm" {
				os.Exit(0)
			}
		}
		askAndExecuteRule(runtimeData, runtimeData.GuiMode)
	},
}

func getInitialRuntimeData(cmd *cobra.Command) shared.ImportLoopData {
	detach(cmd)
	isExecutedByGui, err := cmd.Flags().GetBool("gui-child-mode")
	if err != nil {
		isExecutedByGui = true
	}

	var infoApi shared.InfoInterface
	var dbusConn *dbus.Conn

	if isExecutedByGui {
		dbusConn, err = dbus.SessionBus()
		if err != nil {
			panic(err)
		}

		infoApi = guiApi.InfoApi{
			BusObject: shared.DBusObject(dbusConn),
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
		DbusConn:       dbusConn,
		GuiMode:        isExecutedByGui,
	}
}

func detach(cmd *cobra.Command) {
	isDetached, err := cmd.Flags().GetBool("detached")
	if err != nil {
		panic(err)
	}

	if isDetached {
		return
	}
	args := append(os.Args, "--detached")

	command := exec.Command("nohup", args...)

	var stdout bytes.Buffer
	outWriter := io.MultiWriter(os.Stdout, &stdout)
	command.Stdout = outWriter

	var stderr bytes.Buffer
	errWriter := io.MultiWriter(os.Stdout, &stderr)
	command.Stderr = errWriter

	err = command.Run()
	if err != nil {
		panic("failed to start spito: " + err.Error())
	}
	os.Exit(0)
}

func panicIfEnvironment(runtimeData shared.ImportLoopData, ruleConf *shared.RuleConfigLayout, rulesetIdentifier, ruleName string) {
	if ruleConf.Environment {
		runtimeData.InfoApi.Error("Rule which you were trying to check is an environment")
		runtimeData.InfoApi.Error("In order to apply environment use command:")
		runtimeData.InfoApi.Error("spito env", rulesetIdentifier, ruleName)

		os.Exit(1)
	}
}
