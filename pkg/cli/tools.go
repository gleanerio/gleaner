package cli

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	//	"fmt"
	"github.com/spf13/cobra"
)

// need a mock config with context maps for when
// a normalized sha of the triples ends up being generated.
// assumes glcon is being run from dir with assets
var vipercontext = []byte(`
context:
  cache: true
contextmaps:
- file: ./assets/schemaorg-current-https.jsonld
  prefix: https://schema.org/
- file: ./assets/schemaorg-current-https.jsonld
  prefix: http://schema.org/
sources:
- sourcetype: sitemap
  name: test
  logo: https://opentopography.org/sites/opentopography.org/files/ot_transp_logo_2.png
  url: https://opentopography.org/sitemap.xml
  headless: false
  pid: https://www.re3data.org/repository/r3d100010655
  propername: OpenTopography
  domain: http://www.opentopography.org/
  active: false
  credentialsfile: ""
  other: {}
  headlesswait: 0
  delay: 0
  IdentifierType: filesha
`)

// gleanerCmd represents the run command
var toolsCmd = &cobra.Command{
	Use:              "tools",
	TraverseChildren: true,
	Short:            "command to execute tools from gleaner and nabu",
	Long: `These are small tools that do things like generate uuids
`, PersistentPreRun: func(cmd *cobra.Command, args []string) {
		//gleanerViperVal no longer init'd at the root level. so create one for running tools.
		if gleanerViperVal == nil {
			gleanerViperVal = viper.New()
			gleanerViperVal.SetConfigType("yaml")
			gleanerViperVal.ReadConfig(bytes.NewBuffer(vipercontext))
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("tools called")
	},
}

func init() {
	rootCmd.AddCommand(toolsCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// gleanerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// gleanerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
