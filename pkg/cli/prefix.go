package cli

import (
	"fmt"
	run "github.com/gleanerio/nabu/pkg"
	"github.com/spf13/cobra"
	"mime"
)

// checkCmd represents the check command
var prefixCmd = &cobra.Command{
	Use:   "prefix ",
	Short: "nabu prefix command",
	Long:  `Load graphs from prefix to triplestore`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nabu prefix called")
		mime.AddExtensionType(".jsonld", "application/ld+json")
		run.NabuPrefix(nabuViperVal)
	},
}

func init() {
	NabuCmd.AddCommand(prefixCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
