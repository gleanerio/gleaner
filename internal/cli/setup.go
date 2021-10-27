package cli

import (
	"fmt"
	"github.com/earthcubearchitecture-project418/gleaner/internal/check"
	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	configTypes "github.com/earthcubearchitecture-project418/gleaner/internal/config"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// setupCmd represents the setup command
var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "setup gleaner process",
	Long:  `connects to S3 store, creates buckets, `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setup called")
		setup(glrVal, cfgPath, cfgName)
	},
}

func init() {
	gleanerCmd.AddCommand(setupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// setupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// setupCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func setup(filename string, cfgPath string, cfgName string) {
	var v1 *viper.Viper
	var err error

	v1, err = configTypes.ReadGleanerConfig(filename, path.Join(cfgPath, cfgName))

	mc := common.MinioConnection(v1)

	// Validate Minio is up  TODO:  validate all expected containers are up
	log.Println("Validating access to object store")
	err = check.ConnCheck(mc)
	if err != nil {
		log.Printf("Connection issue, make sure the minio server is running and accessible. %s ", err)
		os.Exit(1)
	}
	// If requested, set up the buckets
	log.Println("Setting up buckets")
	err = check.MakeBuckets(mc)
	if err != nil {
		log.Println("Error making buckets for setup call")
		os.Exit(1)
	}

	err = check.Buckets(mc)
	if err != nil {
		log.Printf("Can not find bucket. %s ", err)
		os.Exit(1)
	}

	log.Println("Buckets generated.  Object store should be ready for runs")
	os.Exit(0)

}
