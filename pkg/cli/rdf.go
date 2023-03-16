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
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/millers/graph"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// batchCmd represents the batch command
var rdfCmd = &cobra.Command{
	Use:              "rdf",
	TraverseChildren: true,
	Short:            "take a jsonld string and process it through Obj2RDF",
	Long: `Execute gleaner to process a jsonld string,  process it through Obj2RDF.
--jsonld jsonld file to read (reading from stdin, works)

`,

	Run: func(cmd *cobra.Command, args []string) {
		//fmt.Println("uuid called")
		//var runSources []string
		//if sourceVal != "" {
		//	runSources = append(runSources, sourceVal)
		//}
		source := configTypes.Sources{
			Name:             "rdfCmd",
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

		maps := []map[string]interface{}{
			{"file": "./configs/schemaorg-current-https.jsonld", "prefix": "https://schema.org/"},
			{"file": "./configs/schemaorg-current-http.jsonld", "prefix": "http://schema.org/"},
		}
		conf := map[string]interface{}{
			"context": map[string]interface{}{"contextmap": maps},
			"sources": []map[string]interface{}{{"name": "testSource"}},
		}

		var viper = viper.New()
		for key, value := range conf {
			viper.Set(key, value)
		}

		proc, options := common.JLDProc(viper)

		rdfRes, err := graph.Obj2RDF(jsonld, proc, options)
		if err != nil {
			log.Error("ERROR: tools rdf:", "Action: getting rdf:", err)
		}

		log.Info("rdf:", fmt.Sprint(rdfRes))
		fmt.Print("\nrdf:\n", fmt.Sprintf("%v", rdfRes))
	},
}

func init() {
	toolsCmd.AddCommand(rdfCmd)

	// Here you will define your flags and configuration settings.
	rdfCmd.Flags().StringVar(&jsonVal, "jsonld", "", "jsonld file to read")

	log.SetLevel(log.ErrorLevel)
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
