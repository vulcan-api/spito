package cmdNew

import (
	_ "embed"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var NewCmd = &cobra.Command{
	Use:   "new {ruleset name}",
	Short: "Creates new ruleset",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := newRuleset(args[0])
		if err != nil {
			fmt.Println("Failed to create new ruleset")
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
	},
}

//go:embed "initial_ruleset/name-of-rule.lua"
var spitoExampleRule []byte

//go:embed "initial_ruleset/spito.yaml"
var spitoYaml []byte

func newRuleset(directoryName string) error {
	if err := os.Mkdir(directoryName, os.ModePerm); err != nil {
		return err
	}

	err := os.WriteFile(directoryName+"/name-of-rule.lua", spitoExampleRule, os.ModePerm)
	if err != nil {
		return err
	}

	return os.WriteFile(directoryName+"/spito.yaml", spitoYaml, os.ModePerm)
}
