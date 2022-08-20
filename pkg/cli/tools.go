package cli

import (
	"fmt"
	"github.com/spf13/cobra"
)

// gleanerCmd represents the run command
var toolsCmd = &cobra.Command{
	Use:              "tools",
	TraverseChildren: true,
	Short:            "command to execute tools from gleaner and nabu",
	Long: `These are small tools that do things like generate uuids
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gleaner called")
	},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
