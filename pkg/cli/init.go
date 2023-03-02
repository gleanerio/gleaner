package cli

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:              "init",
	TraverseChildren: true,
	Short:            "This initialize a config directory which are used create config files",
	Long: `config init creates template configuration files. :
localConfig.yaml - configuration file for services
sources.csv - a csv listing of sources that are uses to generate lists of files to harvest
gleaner_base.yaml - base configuration file for gleaner
nabu_base. yaml - base configuration for nabu
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("init called")
		err := initCfg(cfgPath, cfgName, configBaseFiles)
		if err != nil {
			log.Error(err)
		}

	},
}

func init() {
	configCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initCfg(cfgpath string, cfgName string, configBaseFiles map[string]string) error {
	fmt.Println("make a config template is there isn't already one")
	var basePath = path.Join(cfgpath, cfgName)
	if _, err := os.Stat(basePath); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(basePath, os.ModePerm)
		if err != nil {
			log.Error(err)
		}
	}

	// do not overwrite the source.csv or servers.yaml
	_, err := os.Stat(path.Join(cfgpath, cfgName, configBaseFiles["sources"]))
	if err == nil {
		copy(path.Join(cfgpath, "template", configBaseFiles["sources"]), path.Join(cfgpath, cfgName, configBaseFiles["sources"]+"_latest"))
		delete(configBaseFiles, "sources")
	}
	_, err = os.Stat(path.Join(cfgpath, cfgName, configBaseFiles["servers"]))
	if err == nil {
		copy(path.Join(cfgpath, "template", configBaseFiles["servers"]), path.Join(cfgpath, cfgName, configBaseFiles["servers"]+"_latest"))
		delete(configBaseFiles, "servers")
	}
	if err == nil {
		var configdoc = path.Join(cfgpath, cfgName, configBaseFiles["configdoc"])
		var doc = path.Join("docs", configBaseFiles["configdoc"])
		copy(doc, configdoc)
		delete(configBaseFiles, "configdoc")
	}
	// copy files listed in config.go: configBaseFiles
	for _, f := range configBaseFiles {
		var template = path.Join(cfgpath, cfgName, f)
		var config = path.Join("configs", "template", f)
		copy(config, template)
	}
	//DownloadFile(path.Join(cfgpath, cfgName, "schemaorg-current-https.jsonld"), "https://schema.org/version/latest/schemaorg-current-https.jsonld")
	// just use a common one, else the config base needed to be patched every time
	DownloadFile(path.Join(cfgpath, "schemaorg-current-https.jsonld"), "https://schema.org/version/latest/schemaorg-current-https.jsonld")
	return nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
