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
	"fmt"
	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	configTypes "github.com/earthcubearchitecture-project418/gleaner/internal/config"
	"github.com/earthcubearchitecture-project418/gleaner/internal/millers"
	"github.com/earthcubearchitecture-project418/gleaner/internal/organizations"
	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner"
	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner/acquire"
	"log"
	"path"

	"github.com/spf13/cobra"
)

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Execute gleaner process",
	Long: `run gleaner process to extract JSON-LD from pages using sitemaps, conver to triples
and store to a S3 server:
--cfgName
--mode`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("batch called")
		cli(glrVal, cfgPath, cfgName, modeVal)
	},
}

func init() {
	gleanerCmd.AddCommand(batchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func cli(filename string, cfgPath string, cfgName string, mode string) {

	v1, err := configTypes.ReadGleanerConfig(filename, path.Join(cfgPath, cfgName))
	if err != nil {
		panic(err)
	}
	mc := common.MinioConnection(v1)

	mcfg := v1.GetStringMapString("gleaner")

	// Build the org graph
	// err := organizations.BuildGraphMem(mc, v1) // parfquet testing
	err = organizations.BuildGraph(mc, v1)
	if err != nil {
		log.Print(err)
	}

	// Index the sitegraphs first, if any
	fn, err := acquire.GetGraph(mc, v1)
	if err != nil {
		log.Print(err)
	}
	log.Println(fn)

	// If configured, summon sources
	if mcfg["summon"] == "true" {
		summoner.Summoner(mc, v1)
	}

	// if configured, process summoned sources fronm JSON-LD to RDF (nq)
	if mcfg["mill"] == "true" {
		millers.Millers(mc, v1) // need to remove rundir and then fix the compile
	}
}