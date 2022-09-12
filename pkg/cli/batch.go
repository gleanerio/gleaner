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
	"github.com/gleanerio/gleaner/internal/common"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/pkg"
	bolt "go.etcd.io/bbolt"
	"os"

	log "github.com/sirupsen/logrus"
	"path"

	"github.com/spf13/cobra"
)

var sourceVal string
var summonVal bool
var millVal bool

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:              "batch",
	TraverseChildren: true,
	Short:            "Execute gleaner process",
	Long: `run gleaner process to extract JSON-LD from pages using sitemaps, conver to triples
and store to a S3 server:
--cfgName
--mode`,

	Run: func(cmd *cobra.Command, args []string) {
		log.Println("batch called")

		var runSources []string
		if sourceVal != "" {
			runSources = append(runSources, sourceVal)
		}
		Batch(glrVal, cfgPath, cfgName, modeVal, runSources)
	},
}

func init() {
	gleanerCmd.AddCommand(batchCmd)

	// Here you will define your flags and configuration settings.
	batchCmd.Flags().StringVar(&sourceVal, "source", "", "Override config file source(s) to specify an index target")
	batchCmd.Flags().StringVar(&modeVal, "mode", "mode", "Set the mode")
	batchCmd.Flags().BoolVar(&summonVal, "summon", false, "override summon value with True")
	batchCmd.Flags().BoolVar(&millVal, "mill", false, "override mill value with True")
	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// batchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// batchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func Batch(filename string, cfgPath string, cfgName string, mode string, runSources []string) {

	v1, err := configTypes.ReadGleanerConfig(filename, path.Join(cfgPath, cfgName))
	if err != nil {
		panic(err)
	}
	mc := common.MinioConnection(v1)
	// setup the KV store to hold a record of indexed resources
	db, err := bolt.Open(path.Join(cfgPath, cfgName, "gleaner.db"), 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	//var gln = v1.Sub("gleaner")
	gln := v1.GetStringMapString("gleaner")
	if millVal {
		//gln.Set("mill", "true")
		gln["mill"] = "true"
		v1.Set("gleaner", gln)
	}
	if summonVal {
		//gln.Set("summon", "true")
		gln["summon"] = "true"
		v1.Set("gleaner", gln)
	}
	if len(runSources) > 0 {

		v1, err = configTypes.PruneSources(v1, runSources)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
	pkg.Cli(mc, v1, db)
}
