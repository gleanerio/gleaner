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
	"github.com/spf13/cobra"
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
				break
			}
		}
		//fmt.Println(jsonld)
		uuid := common.GetSHA(jsonld)
		fmt.Println("urn:", uuid)
	},
}

func init() {
	gleanerCmd.AddCommand(uuidCmd)

	// Here you will define your flags and configuration settings.
	uuidCmd.Flags().StringVar(&sourceVal, "jsonld", "", "jsonld file to read")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
