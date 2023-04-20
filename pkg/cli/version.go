package cli

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// gleanerCmd represents the run command
var versionCmd = &cobra.Command{
	Use:              "version",
	TraverseChildren: true,
	Short:            "returns version ",
	Long: `returns version
`, PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initGleanerConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("gleaner called")
		log.Info("gleaner called")

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
	log.Println("Version")
}
