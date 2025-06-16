package cli

import (
	"fmt"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile, cfgName, cfgPath string
var gleanerViperVal *viper.Viper
var gleanerName string

// Nabu variables
var viperVal *viper.Viper
var mc *minio.Client
var bucketVal string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "glcon",
	Short: "gleaner and nabu combined CLI",
	Long: `gleaner and nabu combined CLI tool for data extraction and processing

gleaner: Extract JSON-LD from web pages and convert to triples
nabu: Process triples, load to triple stores, and manage data
`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "cfg", "", "config file (default is configs/local)")
	rootCmd.PersistentFlags().StringVar(&cfgPath, "cfgPath", "configs", "base location for config files (default is configs/)")
	rootCmd.PersistentFlags().StringVar(&cfgName, "cfgName", "local", "config file name (default is local so configs/local)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gleaner" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".gleaner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
