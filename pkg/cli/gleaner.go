package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

// gleanerCmd represents the run command
var gleanerCmd = &cobra.Command{
	Use:   "gleaner",
	Short: "command to execute gleaner processes",
	Long: `run gleaner process to extract JSON-LD from pages using sitemaps, conver to triples
and store to a S3 server:
--cfgName
--mode
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gleaner called")
	},
}
var sourceVal, modeVal string

func init() {
	rootCmd.AddCommand(gleanerCmd)

	// Here you will define your flags and configuration settings.
	gleanerCmd.Flags().StringVar(&modeVal, "mode", "mode", "Set the mode")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
