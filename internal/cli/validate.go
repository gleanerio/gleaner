package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "(not implemented)validate gleaner.io files",
	Long:  `(not implemented)This should read and validate the gleaner and nabu files`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("validate called")
		fmt.Println("Not Implemented")
	},
}

func init() {
	configCmd.AddCommand(validateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// validateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// validateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
