/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cli

import (
	"bufio"
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var jsonVal string

// batchCmd represents the batch command
var uuidCmd = &cobra.Command{
	Use:              "uuid",
	TraverseChildren: true,
	Short:            "Execute gleaner to get a uuid for a jsonld string",
	Long: `run gleaner process  generate uuid for a JSON-LD 
--cfgName
--mode`,

	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("uuid called")
		var runSources []string
		if sourceVal != "" {
			runSources = append(runSources, sourceVal)
		}

		jsonld := ""

		reader := bufio.NewReader(cmd.InOrStdin())
		for {
			text, err := reader.ReadString('\n')
			jsonld = jsonld + text
			if err != nil {
				if err.Error() != "EOF" {
					log.Error("error:", err)
					os.Exit(1)
				}
				break
			}
		}
		log.Info(jsonld)
		//uuid := common.GetSHA(jsonld)
		uuid, err := common.GetNormSHA(jsonld, gleanerViperVal) // Moved to the normalized sha value
		if err != nil {
			log.Error("ERROR: uuid generator:", "Action: Getting normalized sha  Error:", err)
		}
		log.Info("urn:", uuid)
		fmt.Println("\nurn:", uuid)
	},
}

func init() {
	toolsCmd.AddCommand(uuidCmd)

	// Here you will define your flags and configuration settings.
	uuidCmd.Flags().StringVar(&sourceVal, "jsonld", "", "jsonld file to read")
	log.SetLevel(log.ErrorLevel)
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
