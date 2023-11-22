package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "spito",
	Short: "spito is powerful config management system",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Try running subcommand instead")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(checkFileCmd)
	rootCmd.AddCommand(checkCmd)

	rootCmd.Flags().Bool("gui-child-mode", false, "Tells app that it is executed by gui")
}
