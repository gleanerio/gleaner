package cli

import (
	"errors"
	"fmt"
	"github.com/gleanerio/gleaner/internal/check"
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

// gleanerCmd represents the run command
var gleanerCmd = &cobra.Command{
	Use:              "gleaner",
	TraverseChildren: true,
	Short:            "command to execute gleaner processes",
	Long: `run gleaner process to extract JSON-LD from pages using sitemaps, conver to triples
and store to a S3 server:
--cfgName
--mode
`, PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initGleanerConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("gleaner called")
	},
}
var modeVal string

func init() {

	rootCmd.AddCommand(gleanerCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initGleanerConfig() {

	gleanerViperVal = viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		gleanerViperVal.SetConfigFile(cfgFile)
		gleanerViperVal.SetConfigType("yaml")
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".gleaner" (without extension).
		gleanerViperVal.AddConfigPath(home)
		gleanerViperVal.AddConfigPath(path.Join(cfgPath, cfgName))
		gleanerViperVal.SetConfigType("yaml")
		gleanerViperVal.SetConfigName("gleaner")
	}
	// If a config file is found, read it in.
	if err := gleanerViperVal.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using gleaner config file:", gleanerViperVal.ConfigFileUsed())
	} else {
		log.Fatal("error reading config file ", err)
	}

	// some checks
	mc := common.MinioConnection(gleanerViperVal)
	err := check.ConnCheck(mc)
	if err != nil {
		err = errors.New(err.Error() + fmt.Sprintf(" Ignore that. It's not the bucket. check config/minio: address, port, ssl. connection info: endpoint: %v ", mc.EndpointURL()))
		log.Fatal("cannot connect to minio: ", err)
	}

	bucketVal, err = config.GetBucketName(gleanerViperVal)
	if err != nil {
		log.Fatal("cannot read bucketname from : %s ", err)
	}
}
