package cli

import (

	//	"fmt"
	"errors"
	"fmt"
	"github.com/gleanerio/gleaner/internal/check"
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/config"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		log.Info("gleaner called")

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
	// gleanerViperVal is declared in cli/root.go
	var err error
	if cfgFile != "" {

		dir, base := path.Split(cfgFile)
		gleanerViperVal, err = configTypes.ReadGleanerConfig(base, dir)
		if err != nil {
			//panic(err)
			fmt.Println("cannot find config file. Did you 'glcon generate --cfgName XXX' ")
			log.Fatal("cannot find config file. Did you 'glcon generate --cfgName XXX' ")
			os.Exit(66)
		}
	} else {

		gleanerViperVal, err = configTypes.ReadGleanerConfig(gleanerName, path.Join(cfgPath, cfgName))
		if err != nil {
			//panic(err)
			fmt.Println("cannot find config file. Did you 'glcon generate --cfgName XXX' ")
			log.Fatal("cannot find config file. Did you 'glcon generate --cfgName XXX' ")
			os.Exit(66)
		}
	}

	// some checks
	mc := common.MinioConnection(gleanerViperVal)
	err = check.ConnCheck(mc)
	if err != nil {
		err = errors.New(err.Error() + fmt.Sprintf(" Ignore that. It's not the bucket. check config/minio: address, port, ssl. connection info: endpoint: %v ", mc.EndpointURL()))
		log.Fatal("cannot connect to minio: ", err)
	}

	bucketVal, err = config.GetBucketName(gleanerViperVal)
	if err != nil {
		log.Fatal("cannot read bucketname from : ", err)
	}
}
