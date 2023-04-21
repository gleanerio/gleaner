package cli

import (
	"fmt"
	"github.com/gleanerio/gleaner/pkg"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// gleanerCmd represents the run command
var versionCmd = &cobra.Command{
	Use:              "version",
	TraverseChildren: true,
	Short:            "returns version ",
	Long: `returns version
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("version called")

		returnVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func returnVersion() {
	fmt.Println("Version: " + pkg.VERSION)
}
