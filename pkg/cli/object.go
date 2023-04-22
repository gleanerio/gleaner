package cli

import (
	"fmt"
	run "github.com/gleanerio/nabu/pkg"
	"github.com/spf13/cobra"
	"mime"
)

var objectVal string

// checkCmd represents the check command
var objectCmd = &cobra.Command{
	Use:              "object [flags] configName",
	TraverseChildren: true,
	Short:            "nabu object command",
	Long:             `Load a graph object to triplestore`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nabu object called")
		mime.AddExtensionType(".jsonld", "application/ld+json")
		run.NabuObject(nabuViperVal, bucketVal, objectVal)
	},
}

func init() {
	NabuCmd.AddCommand(prefixCmd)

	// Here you will define your flags and configuration settings.
	NabuCmd.Flags().StringVar(&objectVal, "object", "", "object")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
