package cli

import (
	"fmt"
	run "github.com/gleanerio/nabu/pkg"
	"github.com/spf13/cobra"
	"mime"
)

// checkCmd represents the check command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "nabu prune command",
	Long:  `Prune graphs in triplestore not in objectVal store`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nabu prune called")
		mime.AddExtensionType(".jsonld", "application/ld+json")
		run.NabuPrune(nabuViperVal)
	},
}

func init() {
	NabuCmd.AddCommand(pruneCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
