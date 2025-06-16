package cli

import (
	"errors"
	"fmt"
	"mime"
	"os"
	"path"
	"path/filepath"

	"github.com/gleanerio/gleaner/nabu/pkg/config"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/common"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/objects"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile, cfgURL, cfgName, cfgPath, nabuConfName string
var minioVal, portVal, accessVal, secretVal, bucketVal string
var sslVal, dangerousVal bool
var viperVal *viper.Viper
var mc *minio.Client
var prefixVal, endpointVal string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "nabu",
	Short: "nabu ",
	Long: `nabu
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
	//LOG_FILE := "nabu.log" // log to custom file
	//logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	//if err != nil {
	//log.Panic(err)
	//return
	//}
	////defer logFile.Close()

	//log.SetOutput(logFile) // Set log out put and enjoy :)

	//log.SetFlags(log.Lshortfile | log.LstdFlags) // optional: log date-time, filename, and line number
	//log.Println("Logging to custom file")
	//log.Println("EarthCube Nabu")
	common.InitLogging()

	mime.AddExtensionType(".jsonld", "application/ld+json")

	akey := os.Getenv("MINIO_ACCESS_KEY")
	skey := os.Getenv("MINIO_SECRET_KEY")
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&prefixVal, "prefix", "", "prefix to run. use source in future.")
	// This needs to be done right... there are prov/source milled/source
	// will need a custom validator to say, hey use prefix.
	//	rootCmd.PersistentFlags().StringVar(&prefixVal, "source", "", "prefix to run. Consistency with glcon commend")

	// Enpoint Server setting var
	rootCmd.PersistentFlags().StringVar(&endpointVal, "endpoint", "", "end point server set for the SPARQL endpoints")

	rootCmd.PersistentFlags().StringVar(&cfgURL, "cfgURL", "configs", "URL location for config file")
	rootCmd.PersistentFlags().StringVar(&cfgPath, "cfgPath", "configs", "base location for config files (default is configs/)")
	rootCmd.PersistentFlags().StringVar(&cfgName, "cfgName", "local", "config file (default is local so configs/local)")
	rootCmd.PersistentFlags().StringVar(&nabuConfName, "nabuConfName", "nabu", "config file (default is local so configs/local)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "cfg", "", "compatibility/overload: full path to config file (default location gleaner in configs/local)")

	// minio env variables
	rootCmd.PersistentFlags().StringVar(&minioVal, "address", "localhost", "FQDN for server")
	rootCmd.PersistentFlags().StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	rootCmd.PersistentFlags().StringVar(&accessVal, "access", akey, "Access Key ID")
	rootCmd.PersistentFlags().StringVar(&secretVal, "secret", skey, "Secret access key")
	rootCmd.PersistentFlags().StringVar(&bucketVal, "bucket", "gleaner", "The configuration bucket")

	rootCmd.PersistentFlags().BoolVar(&sslVal, "ssl", false, "Use SSL boolean")
	rootCmd.PersistentFlags().BoolVar(&dangerousVal, "dangerous", false, "Use dangerous mode boolean")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var err error
	//viperVal := viper.New()
	if cfgFile != "" {
		// Use config file from the flag.
		//viperVal.SetConfigFile(cfgFile)
		viperVal, err = config.ReadNabuConfig(filepath.Base(cfgFile), filepath.Dir(cfgFile))
		if err != nil {
			log.Fatal("cannot read config %s", err)
		}
	} else if cfgURL != "" {
		viperVal, err = config.ReadNabuConfigURL(cfgURL)
		if err != nil {
			log.Fatal("cannot read config URL %s", err)
		}
	} else {
		// Find home directory.
		//home, err := os.UserHomeDir()
		//cobra.CheckErr(err)
		//
		//// Search config in home directory with name "nabu" (without extension).
		//viperVal.AddConfigPath(home)
		//viperVal.AddConfigPath(path.Join(cfgPath, cfgName))
		//viperVal.SetConfigType("yaml")
		//viperVal.SetConfigName("nabu")
		viperVal, err = config.ReadNabuConfig(nabuConfName, path.Join(cfgPath, cfgName))
		if err != nil {
			log.Fatal("cannot read config %s", err)
		}
	}

	//viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.

	mc, err = objects.MinioConnection(viperVal)
	if err != nil {
		log.Fatal("cannot connect to minio: %s", err)
	}

	err = common.ConnCheck(mc)
	if err != nil {
		err = errors.New(err.Error() + fmt.Sprintf(" Ignore that. It's not the bucket. check config/minio: address, port, ssl. connection info: endpoint: %v ", mc.EndpointURL()))
		log.Fatal("cannot connect to minio: ", err)
	}

	bucketVal, err = config.GetBucketName(viperVal)
	if err != nil {
		log.Fatal("cannot read bucketname from : %s ", err)
	}
	// Override prefix in config if flag set
	//if isFlagPassed("prefix") {
	//	out := viperVal.GetStringMapString("objects")
	//	b := out["bucket"]
	//	p := prefixVal
	//	// r := out["region"]
	//	// v1.Set("objects", map[string]string{"bucket": b, "prefix": NEWPREFIX, "region": r})
	//	viperVal.Set("objects", map[string]string{"bucket": b, "prefix": p})
	//}

	if dangerousVal {
		viperVal.Set("flags.dangerous", true)
	}

	if endpointVal != "" {
		viperVal.Set("flags.endpoint", endpointVal)
	}

	if prefixVal != "" {
		//out := viperVal.GetStringMapString("objects")
		//d := out["domain"]

		var p []string
		p = append(p, prefixVal)

		viperVal.Set("objects.prefix", p)

		//p := prefixVal
		// r := out["region"]
		// v1.Set("objects", map[string]string{"bucket": b, "prefix": NEWPREFIX, "region": r})
		//viperVal.Set("objects", map[string]string{"domain": d, "prefix": p})
	}

}
