package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"

	"github.com/spf13/viper"
)

var cfgFile, cfgName, cfgPath string
var minioVal, portVal, accessVal, secretVal, bucketVal string
var sslVal bool
var viperVal *viper.Viper

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gleaner",
	Short: "Extracts JSON-LD from web pages exposed in a domains sitemap file. ",
	Long: `The gleaner.io stack harvests JSON-LD from webpages using sitemaps and other tools
store files in S3 (we use minio), uploads triples to be processed by nabu (the next step in the process)
configuration is now focused on a directory (default: configs/local) with will contain the
process to configure and run is:
* gleaner config init --cfgName {default:local}
  edit files, servers.yaml, sources.csv
* gleaner config generate --cfgName  {default:local}
* gleaner run setup --cfgName  {default:local}
* gleaner run  --cfgName  {default:local}
* run nabu (better description)
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {

	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	log.Println("EarthCube Gleaner")
	akey := os.Getenv("MINIO_ACCESS_KEY")
	skey := os.Getenv("MINIO_SECRET_KEY")
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgPath, "cfgPath", "configs", "base location for config files (default is configs/)")
	rootCmd.PersistentFlags().StringVar(&cfgName, "cfgName", "local", "config file (default is local so configs/local)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "cfg", "", "compatibility/overload: full path to config file (default location gleaner in configs/local)")

	// minio env variables
	rootCmd.PersistentFlags().StringVar(&minioVal, "address", "localhost", "FQDN for server")
	rootCmd.PersistentFlags().StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	rootCmd.PersistentFlags().StringVar(&accessVal, "access", akey, "Access Key ID")
	rootCmd.PersistentFlags().StringVar(&secretVal, "secret", skey, "Secret access key")
	rootCmd.PersistentFlags().StringVar(&bucketVal, "bucket", "gleaner", "The configuration bucket")

	rootCmd.PersistentFlags().BoolVar(&sslVal, "ssl", false, "Use SSL boolean")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viperVal := viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		viperVal.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gleaner" (without extension).
		viperVal.AddConfigPath(home)
		viperVal.AddConfigPath(path.Join(cfgPath, cfgName))
		viperVal.SetConfigType("yaml")
		viperVal.SetConfigName("gleaner")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viperVal.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viperVal.ConfigFileUsed())
	}

}
