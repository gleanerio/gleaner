package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate gleaner.io config files from a directory that has been intialized",
	Long: `Generate creates config files for the gleaner.io tools (gleaner and nabu). Before running command 
run 
# gleaner config init --confName {default: local}

Usually you will need to edit servers.yaml and sources.csv.
A copy of the files (one per DAY) is saved.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generate called")
	},
}

func init() {
	configCmd.AddCommand(generateCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
