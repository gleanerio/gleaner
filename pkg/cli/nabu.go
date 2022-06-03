package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var nabuConfig *viper.Viper

// gleanerCmd represents the run command
var NabuCmd = &cobra.Command{
	Use:   "nabu",
	Short: "command to execute nabu processes",
	Long: `run naub process to upload triples and prune results
--cfgName
--mode
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("nabu called")
	},
}

var prefixVal string

func init() {
	rootCmd.AddCommand(NabuCmd)

	// Here you will define your flags and configuration settings.
	//NabuCmd.Flags().StringVar(&nabuVal, "cfg", "nabu", "Configuration file")
	NabuCmd.Flags().StringVar(&prefixVal, "prefix", "", "Prefix to override config file setting")
	NabuCmd.Flags().MarkDeprecated("source", "use --prefix prov/source or milled/source to override loading")
	NabuCmd.Flags().MarkShorthandDeprecated("source", "use --prefix prov/source or milled/source to override loading")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
