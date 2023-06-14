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
	"github.com/gleanerio/gleaner/internal/common"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var jsonVal string
var idTypeVal string
var idPathVal string // string separated by a comman

// need a mock config with context maps for when
// a normalized sha of the triples ends up being generated.
// assumes glcon is being run from dir with assets
//var vipercontext = []byte(`
//context:
//  cache: true
//contextmaps:
//- file: ./assets/schemaorg-current-https.jsonld
//  prefix: https://schema.org/
//- file: ./assets/schemaorg-current-https.jsonld
//  prefix: http://schema.org/
//sources:
//- sourcetype: sitemap
//  name: test
//  logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
//  url: https://opentopography.org/sitemap.xml
//  headless: false
//  pid: https://www.re3data.org/repository/r3d100010655
//  propername: OpenTopography
//  domain: http://www.opentopography.org/
//  active: false
//  credentialsfile: ""
//  other: {}
//  headlesswait: 0
//  delay: 0
//  IdentifierType: filesha
//`)

// batchCmd represents the batch command
var identifierCmd = &cobra.Command{
	Use:              "id",
	TraverseChildren: true,
	Short:            "Generate the identifier a jsonld string",
	Long: `Execute gleaner to generate the identifier a jsonld string. 
--jsonld jsonld file to read (reading from stdin, works)
--idtype (filesha |identifiersha | identifierstring )
--idPath (json path rule for the identifier)

There are three types of idtype: 
* filesha -just generate the sha for the entire jsonld file
* identifiersha - use identifieter rules to determine the identifier, and generate the sha
* identifierstring -(not yet implemented) use idenfitfer rules to determine identifier, then convert to a url safe string.
`,

	Run: func(cmd *cobra.Command, args []string) {

		// need a  config with context maps for when
		// a normalized sha of the triples ends up being generated.
		// there is a mock that assumes glcon is being run from dir with assets,
		// but that is not always the case.
		var err error

		if gleanerViperVal == nil {
			gleanerViperVal = viper.New()
			gleanerViperVal.SetConfigType("yaml")
			gleanerViperVal.ReadConfig(bytes.NewBuffer(vipercontext))
		}

		source := configTypes.Sources{
			Name:             "identifierCmd",
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

		identifier, err := common.GenerateIdentifier(gleanerViperVal, source, jsonld)
		if err != nil {
			log.Error("ERROR: uuid generator:", "Action: Getting normalized sha  Error:", err)
		}
		log.Info("urn:", fmt.Sprint(identifier))
		fmt.Println("\nurn:", fmt.Sprint(identifier))
	},
}

func init() {
	toolsCmd.AddCommand(identifierCmd)

	// Here you will define your flags and configuration settings.
	identifierCmd.Flags().StringVar(&jsonVal, "jsonld", "", "jsonld file to read")
	identifierCmd.Flags().StringVar(&idTypeVal, "idtype", "", "identifiertype to generate")
	identifierCmd.Flags().StringVar(&idPathVal, "idtPath", "", "id path to use")
	log.SetLevel(log.ErrorLevel)
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
