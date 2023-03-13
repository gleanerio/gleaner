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

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:              "check",
	TraverseChildren: true,
	Short:            "(not implemented)Check the connectivity to the Minio server",
	Long:             `(not implemented)This will read the configs/{cfgPath}/gleaner file, and try to connect to the minio server`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("check called")
		doCheck(glrVal, cfgPath, cfgName)
	},
}

func init() {
	gleanerCmd.AddCommand(checkCmd)
	configCmd.AddCommand(checkCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func doCheck(filename string, cfgPath string, cfgName string) {
	var v1 *viper.Viper
	var err error

	v1, err = configTypes.ReadGleanerConfig(filename, path.Join(cfgPath, cfgName))
	if err != nil {
		log.Fatal("Error reading gleaner config. Did you 'glcon generate --cfgName XXX'", err)
		os.Exit(66)
	}

	mc := common.MinioConnection(v1)

	err = check.PreflightChecks(mc, v1)
	if err != nil {
		log.Fatal("Failed Check", err)
		os.Exit(66)
	}
	fmt.Println("Check successful: ")
}
