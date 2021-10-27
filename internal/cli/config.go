package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "commands to intialize, and generate tools: gleaner and nabu",
	Long: `This command is used to initialize configuration files for the gleaner.io ecosystem. 
gleaner harvests and converts JSON-LD files from sites. 
nabu uploads and manages data processed by gleaner to a sparql triplestore
* sites are stored in a sources.csv
* configs/{localdirector}/servers.yaml is the configuration files for servers.
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("config called")
	},
}

var glrVal, nabuVal, sourcesVal, templateGleaner, templateNabu string

var configBaseFiles = map[string]string{"gleaner": "gleaner_base.yaml", "sources": "sources.csv",
	"nabu": "nabu_base.yaml", "servers": "servers.yaml", "readme": "readme.txt"}

var gleanerFileNameBase = "gleaner"
var nabuFilenameBase = "nabu"

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().StringVar(&templateGleaner, "template_gleaner", configBaseFiles["gleaner"], "Configuration Template or Cofiguration file")
	configCmd.PersistentFlags().StringVar(&templateNabu, "template_nabu", configBaseFiles["nabu"], "Configuration Template or Cofiguration file")
	configCmd.PersistentFlags().StringVar(&glrVal, "gleaner", gleanerFileNameBase+".yaml", "output gleaner file to")
	configCmd.PersistentFlags().StringVar(&nabuVal, "nabu", nabuFilenameBase+".yaml", "output nabu file to")
	configCmd.PersistentFlags().StringVar(&sourcesVal, "sourcemaps", "sources.csv", "sources as csv")
	rootCmd.MarkPersistentFlagRequired("cfgName")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
