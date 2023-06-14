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
	"bytes"
	"fmt"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// batchCmd represents the batch command
var jsonLdCmd = &cobra.Command{
	Use:              "jsonld",
	TraverseChildren: true,
	Short:            "take a jsonld string and process it through the context",
	Long: `Execute gleaner to process a jsonld string, and run through context and other
processing.
--jsonld jsonld file to read (reading from stdin, works)
--idtype (filesha |identifiersha | identifierstring )
--idPath (json path rule for the identifier)

There are three types of idtype: 
* filesha -just generate the sha for the entire jsonld file
* identifiersha - use identifieter rules to determine the identifier, and generate the sha
* identifierstring -(not yet implemented) use idenfitfer rules to determine identifier, then convert to a url safe string.
`,

	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("uuid called")
		//var runSources []string
		//if sourceVal != "" {
		//	runSources = append(runSources, sourceVal)
		//}

		if gleanerViperVal == nil {
			gleanerViperVal = viper.New()
			gleanerViperVal.SetConfigType("yaml")
			gleanerViperVal.ReadConfig(bytes.NewBuffer(vipercontext))
		}

		source := configTypes.Sources{
			Name:             "jsonldCmd",
			IdentifierType:   idTypeVal,
			FixContextOption: configTypes.Https,
		}
		if len(idPathVal) > 0 {
			source.IdentifierPath = idPathVal
		}
		jsonld := ""
		if jsonVal != "" {
			text, err := os.ReadFile(jsonVal)
			if err != nil {
				log.Fatal(err)
			}
			jsonld = string(text)
		} else {
			// read from command line
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
		}

		log.Info(jsonld)
		//uuid := common.GetSHA(jsonld)
		//uuid, err := common.GetNormSHA(jsonld, gleanerViperVal) // Moved to the normalized sha value
		jsonRes, identifierRes, err := acquire.ProcessJson(gleanerViperVal, &source, "http://example.com/", jsonld)
		if err != nil {
			log.Error("ERROR: json ldtest:", "Action: getting json:", err)
		}
		log.Info("urn:", fmt.Sprint(identifierRes))
		fmt.Println("\nurn:", fmt.Sprint(identifierRes))
		log.Info("jsonld:", fmt.Sprint(jsonRes))
		fmt.Println("\njsonld:", fmt.Sprint(jsonRes))
	},
}

func init() {
	toolsCmd.AddCommand(jsonLdCmd)

	// Here you will define your flags and configuration settings.
	jsonLdCmd.Flags().StringVar(&jsonVal, "jsonld", "", "jsonld file to read")
	jsonLdCmd.Flags().StringVar(&idTypeVal, "idtype", "", "identifiertype to generate")
	jsonLdCmd.Flags().StringVar(&idPathVal, "idtPath", "", "id path to use")
	log.SetLevel(log.ErrorLevel)
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
