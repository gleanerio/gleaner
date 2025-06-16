package cli

import (
	"fmt"
	"github.com/gleanerio/gleaner/nabu/pkg"
	log "github.com/sirupsen/logrus"
	"os"

	"github.com/spf13/cobra"
)

//var bucket, objectVal string

// checkCmd represents the check command
var objectCmd = &cobra.Command{
	Use:   "object",
	Short: "nabu object command",
	Long:  `Load graph object to triplestore`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("object called")
		if len(args) > 0 {
			objectVal := args[0]

			err := pkg.Object(viperVal, mc, bucketVal, objectVal)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			os.Exit(0)
		} else {
			log.Fatal("must have an argument nabu object objectId")
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(objectCmd)

	// Here you will define your flags and configuration settings.
	// bucketVal is available at top level
	//objectCmd.Flags().StringVar(&objectVal, "object", "", "object to load")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
