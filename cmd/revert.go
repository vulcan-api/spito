package cmd

import (
	"fmt"
	"github.com/avorty/spito/internal/checker"
	"github.com/avorty/spito/pkg/vrct/vrctFs"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

var revertCmd = &cobra.Command{
	Use:   "revert {revert number}",
	Short: "Reverts changes by replacing them with backup",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		revertSteps, err := vrctFs.NewRevertSteps()
		if err != nil {
			handleError(err)
		}

		arg := strings.TrimSpace(args[0])
		revertNum, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Println("Failed to parse input, revert number needs to be an integer")
		}

		importLoopData := getInitialRuntimeData(cmd)

		err = revertSteps.Deserialize(revertNum)
		handleError(err)

		err = revertSteps.Apply(checker.GetRevertRuleFn(importLoopData.InfoApi))
		handleError(err)
	},
}
