package config

import (
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

// as read from csv
type Sources struct {
	SourceType string `default:"sitemap"`
	Name       string
	Logo       string
	URL        string
	Headless   bool
	PID        string
	ProperName string
	Domain     string
	Active     bool `default:"true"`
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
		"sourcetype": "sitemap",
		"name":       "",
		"url":        "",
		"logo":       "",
		"headless":   "",
		"pid":        "",
		"propername": "",
		"domain":     "",
	},
}

func populateDefaults(s Sources) Sources {
	if s.SourceType == "" {
		s.SourceType = "sitemap"
	}
	// fix issues, too. Space from CSV causing url errors
	s.URL = strings.TrimSpace(s.URL)
	return s

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

	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		//return csv.NewReader(in)
		return gocsv.LazyCSVReader(in) // Allows use of quotes in CSV
	})

	if err := gocsv.Unmarshal(f, &sources); err != nil {
		fmt.Println("error:", err)
	}

	for i, u := range sources {
		sources[i] = populateDefaults(u)
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
	for i, s := range cfg {
		cfg[i] = populateDefaults(s)
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

func SourceToNabuPrefix(sources []Sources, includeProv bool) []string {

	var prefixes []string
	for _, s := range sources {

		switch s.SourceType {

		case "sitemap":
			prefixes = append(prefixes, "milled/"+s.Name)
			if includeProv {
				prefixes = append(prefixes, "prov/"+s.Name)
			}

		case "sitegraph":
			prefixes = append(prefixes, "summoned/"+s.Name)
			if includeProv {
				prefixes = append(prefixes, "prov/"+s.Name)
			}

		}
	}
	return prefixes
}
