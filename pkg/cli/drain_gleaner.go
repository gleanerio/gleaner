package cli

import (
	"github.com/gleanerio/gleaner/nabu/pkg"

	"os"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var DrainCmd = &cobra.Command{
	Use:   "drain ",
	Short: "nabu drain command",
	Long:  `Remove all objects from a S3 bucket - prefix `,
	Run: func(cmd *cobra.Command, args []string) {
		err := pkg.Drain(viperVal, mc)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(DrainCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
