package cli

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/check"
	"github.com/gleanerio/gleaner/internal/common"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// setupCmd represents the Setup command
var setupCmd = &cobra.Command{
	Use:              "setup",
	TraverseChildren: true,
	Short:            "setup gleaner process",
	Long:             `connects to S3 store, creates buckets, `,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setup called")
		setup(glrVal, cfgPath, cfgName)
	},
}

func init() {
	gleanerCmd.AddCommand(setupCmd)
	configCmd.AddCommand(setupCmd)
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
	if err != nil {
		log.Fatal("Error reading gleaner config. Did you 'glcon generate --cfgName XXX'", err)
		os.Exit(66)
	}

	mc := common.MinioConnection(v1)

	check.Setup(mc, v1)
}
