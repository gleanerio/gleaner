package config

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

// as read from csv
type Sources struct {
	SourceType string
	Name       string
	Logo       string
	URL        string
	Headless   bool
	PID        string
	ProperName string
	Domain     string
	Active     bool
	// SitemapFormat string
	// Active        bool
}

// add needed for file
type SourcesConfig struct {
	Name       string
	Logo       string
	URL        string
	Headless   bool
	PID        string
	ProperName string
	Domain     string
	// SitemapFormat string
	// Active        bool
}

var SourcesTemplate = map[string]interface{}{
	"sources": map[string]string{
		"sourcetype": "",
		"name":       "",
		"url":        "",
		"logo":       "",
		"headless":   "",
		"pid":        "",
		"propername": "",
		"domain":     "",
	},
}

func ReadSourcesCSV(filename string, cfgPath string) ([]Sources, error) {
	var sources []Sources
	var err error
	var fn = path.Join(cfgPath, filename)
	f, err := os.Open(fn)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	if err := gocsv.Unmarshal(f, &sources); err != nil {
		fmt.Println("error:", err)
	}

	for _, u := range sources {
		fmt.Printf("%+v\n", u)
	}
	return sources, err

}

// use full gleaner viper. v1.Sub("sources") fails because it is an array.
// If we need to override with env variables, then we might need to grab this patch https://github.com/spf13/viper/pull/509/files

func ParseSourcesConfig(g1 *viper.Viper) ([]Sources, error) {
	var subtreeKey = "sources"
	var cfg []Sources
	//for key, value := range SourcesTemplate {
	//	g1.SetDefault(key, value)
	//}

	//g1.AutomaticEnv()
	// config already read. substree passed
	err := g1.UnmarshalKey(subtreeKey, &cfg)
	if err != nil {
		panic(fmt.Errorf("error when parsing %v config: %v", subtreeKey, err))
	}
	return cfg, err
}

func GetSourceByType(sources []Sources, key string) []Sources {
	var sourcesSlice []Sources
	for _, s := range sources {
		if s.SourceType == key {
			sourcesSlice = append(sourcesSlice, s)
		}
	}
	return sourcesSlice
}
