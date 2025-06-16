package cli

import (
	"fmt"
	"os"

	"github.com/gleanerio/gleaner/nabu/pkg"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var meiliCmd = &cobra.Command{
	Use:   "meili",
	Short: "nabu meili command",
	Long:  `This will read the configs/{cfgPath}/gleaner and connect and load JSON-LD into MeiliSearch for full text indexing`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("meili called")
		err := pkg.Meili(viperVal, mc)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(meiliCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
