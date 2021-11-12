package cli

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	"github.com/spf13/cobra"
	"log"
)

// initCmd represents the init command
var gdriveCmd = &cobra.Command{
	Use:   "gdrive",
	Short: "This retrieves and store google drive credentials",
	Long: `This retrieves and store google drive credentials
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gdrive called")
		srv, err := acquire.GetDriveCredentials("AIzaSyB7e8ewIkapa2_Eyj-IL9zYP4oG-77Y8J0")
		//srv, err := googledrive.GetDriveCredentials("")
		if err != nil {
			fmt.Println(err)
		}
		log.Println(srv.Drives.List())
		//var folderid = "0B8xif6Jg1upyfnpORzlTZ01kSUpFejFySFBZVlBtaHNWLVlic0QtVzYxQ21lYzNmVk0yRWc"
		//var folderid ="1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd"
		//var folderid = "0B8xif6Jg1upyfnpORzlTZ01kSUpFejFySFBZVlBtaHNWLVlic0QtVzYxQ21lYzNmVk0yRWc"
		//var folderid = "0B8xif6Jg1upyfnpORzlTZ01kSUpFejFySFBZVlBtaHNWLVlic0QtVzYxQ21lYzNmVk0yRWc"
		var folderid = "1TacUQqjpBbGsPQ8JPps47lBXMQsNBRnd"
		list, err := acquire.GetFileList(srv, folderid, false, folderid)
		log.Println(list)

	},
}

func init() {
	configCmd.AddCommand(gdriveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
