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
		cmd.Parent().PersistentPreRun(cmd.Parent(), args)
		initNabuConfig()
	},
	Run: func(cmd *cobra.Command, args []string) {

		mime.AddExtensionType(".jsonld", "application/ld+json")
		fmt.Println("nabu called")
	},
}

var prefixVal []string
var sparqlEndpointVal string

func init() {
	rootCmd.AddCommand(NabuCmd)

	// Here you will define your flags and configuration settings.
	//NabuCmd.Flags().StringVar(&nabuVal, "cfg", "nabu", "Configuration file")
	NabuCmd.PersistentFlags().StringArrayVar(&prefixVal, "prefix", []string{}, "Prefix to override config file setting")
	NabuCmd.PersistentFlags().MarkDeprecated("source", "use --prefix prov/source or milled/source to override loading")
	NabuCmd.PersistentFlags().MarkShorthandDeprecated("source", "use --prefix prov/source or milled/source to override loading")
	// sparql
	NabuCmd.PersistentFlags().StringVar(&sparqlEndpointVal, "sparqlurl", "", "SPARQL_ENDPOINT")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initNabuConfig() {

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
	// some checks
	mc := common.MinioConnection(nabuViperVal)
	err := check.ConnCheck(mc)
	if err != nil {
		err = errors.New(err.Error() + fmt.Sprintf(" Ignore that. It's not the bucket. check config/minio: address, port, ssl. connection info: endpoint: %v ", mc.EndpointURL()))
		log.Fatal("cannot connect to minio: ", err)
	}

	bucketVal, err = config.GetBucketName(nabuViperVal)
	if err != nil {
		log.Fatal("cannot read bucketname from : ", err)
	}

	if len(prefixVal) > 0 {
		//out := viperVal.GetStringMapString("objects")
		//d := out["domain"]

		var p []string
		for _, pre := range prefixVal {
			p = append(p, pre)
		}

		nabuViperVal.Set("objects.prefix", p)

		//p := prefixVal
		// r := out["region"]
		// v1.Set("objects", map[string]string{"bucket": b, "prefix": NEWPREFIX, "region": r})
		//viperVal.Set("objects", map[string]string{"domain": d, "prefix": p})
	}
}
