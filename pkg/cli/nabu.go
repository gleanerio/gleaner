package cli

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"mime"
	"os"
	"path"
)

var nabuConfig *viper.Viper

// gleanerCmd represents the run command
var NabuCmd = &cobra.Command{
	Use:              "nabu",
	TraverseChildren: true,
	Short:            "command to execute nabu processes",
	Long: `run naub process to upload triples and prune results
--cfgName
--mode
`, PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initNabuConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {

		mime.AddExtensionType(".jsonld", "application/ld+json")
		fmt.Println("nabu called")
	},
}

var prefixVal string

func init() {
	rootCmd.AddCommand(NabuCmd)

	// Here you will define your flags and configuration settings.
	//NabuCmd.Flags().StringVar(&nabuVal, "cfg", "nabu", "Configuration file")
	NabuCmd.PersistentFlags().StringVar(&prefixVal, "prefix", "", "Prefix to override config file setting")
	NabuCmd.PersistentFlags().MarkDeprecated("source", "use --prefix prov/source or milled/source to override loading")
	NabuCmd.PersistentFlags().MarkShorthandDeprecated("source", "use --prefix prov/source or milled/source to override loading")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initNabuConfig() {
	//
	//gleanerViperVal = viper.New()
	//if cfgFile != "" {
	//	// Use config file from the flag.
	//	gleanerViperVal.SetConfigFile(cfgFile)
	//} else {
	//	// Find home directory.
	//	home, err := os.UserHomeDir()
	//	cobra.CheckErr(err)
	//
	//	// Search config in home directory with name ".gleaner" (without extension).
	//	gleanerViperVal.AddConfigPath(home)
	//	gleanerViperVal.AddConfigPath(path.Join(cfgPath, cfgName))
	//	gleanerViperVal.SetConfigType("yaml")
	//	gleanerViperVal.SetConfigName("gleaner")
	//}
	//// If a config file is found, read it in.
	//if err := gleanerViperVal.ReadInConfig(); err == nil {
	//	fmt.Fprintln(os.Stderr, "Using gleaner config file:", gleanerViperVal.ConfigFileUsed())
	//} else {
	//	log.Fatal("error reading config file")
	//}

	nabuViperVal = viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		nabuViperVal.SetConfigFile(cfgFile)
		nabuViperVal.SetConfigType("yaml")
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gleaner" (without extension).
		nabuViperVal.AddConfigPath(home)
		nabuViperVal.AddConfigPath(path.Join(cfgPath, cfgName))
		nabuViperVal.SetConfigType("yaml")
		nabuViperVal.SetConfigName("nabu")
	}
	viper.AutomaticEnv() // read in environment variables that match

	if err := nabuViperVal.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using nabu config file:", nabuViperVal.ConfigFileUsed())
	} else {
		log.Fatal("error reading config file", err)
	}

}
