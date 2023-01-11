package config

import (
	"errors"
	"fmt"
	"github.com/gocarina/gocsv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"

	"github.com/utahta/go-openuri"
	"path"
	"strings"
)

const (
	IdentifierSha     string = "identifiersha"
	JsonSha                  = "jsonsha"
	NormalizedJsonSha        = "normalizedjsonsha"
	IdentifierString         = "identifierstring"
	SourceUrl                = "sourceurl"
)

type ContextOption int64

const (
	Strict ContextOption = iota
	Https
	Http
	//	Array
	//	Object
	StandardizedHttps
	StandardizedHttp
)

func (s ContextOption) String() string {
	switch s {
	case Strict:
		return "strict"
	case Https:
		return "https"
	case Http:
		return "http"
		//	case Array:
		//		return "array"
		//	case Object:
		//		return "object"
	case StandardizedHttps:
		return "standardizedHttps"
	case StandardizedHttp:
		return "standardizedHttp"
	}
	return "unknown"
}

// as read from csv
type Sources struct {
	// Valid values for SourceType: sitemap, sitegraph, csv, googledrive, api, and robots
	SourceType      string `default:"sitemap"`
	Name            string
	Logo            string
	URL             string
	Headless        bool `default:"false"`
	PID             string
	ProperName      string
	Domain          string
	Active          bool                   `default:"true"`
	CredentialsFile string                 // do not want someone's google api key exposed.
	Other           map[string]interface{} `mapstructure:",remain"`
	// SitemapFormat string
	// Active        bool
	HeadlessWait   int    // if loading is slow, wait
	Delay          int64  // A domain-specific crawl delay value
	IdentifierPath string // JSON Path to the identifier
	ApiPageLimit int
	FixContextOption ContextOption
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
	HeadlessWait     int    // is loading is slow, wait
	Delay            int64  // A domain-specific crawl delay value
	IdentifierPath   string // JSON Path to the identifier
	IdentifierType   string
	FixContextOption ContextOption
}

var SourcesTemplate = map[string]interface{}{
	"sources": map[string]string{
		"sourcetype":       "sitemap",
		"name":             "",
		"url":              "",
		"logo":             "",
		"headless":         "",
		"pid":              "",
		"propername":       "",
		"domain":           "",
		"credentialsfile":  "",
		"headlesswait":     "0",
		"delay":            "0",
		"identifierspath":  "",
		"identifiertype":   JsonSha,
		"fixcontextoption": "https",
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
	var fn = ""
	// if it's a url
	if strings.HasPrefix(filename, "https://") || strings.HasPrefix(filename, "http://") {
		fn = filename
	} else if strings.HasPrefix(filename, "/") {
		// its a full path
		fn = filename
	} else {
		fn = path.Join(cfgPath, filename)
	}

	f, err := openuri.Open(fn)
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

func GetSources(g1 *viper.Viper) ([]Sources, error) {
	var subtreeKey = "sources"
	var cfg []Sources
	//for key, value := range SourcesTemplate {
	//	g1.SetDefault(key, value)
	//}

	//g1.AutomaticEnv()
	// config already read. substree passed
	err := g1.UnmarshalKey(subtreeKey, &cfg)
	if err != nil {
		log.Fatal("error when parsing ", subtreeKey, " config: ", err)
		//No sources, so nothing to run
	}
	for i, s := range cfg {
		cfg[i] = populateDefaults(s)
	}
	return cfg, err
}

func GetActiveSources(g1 *viper.Viper) ([]Sources, error) {
	var activeSources []Sources

	sources, err := GetSources(g1)
	if err != nil {
		return nil, err
	}
	for _, s := range sources {
		if s.Active == true {
			activeSources = append(activeSources, s)
		}
	}
	return activeSources, err
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

func GetActiveSourceByType(sources []Sources, key string) []Sources {
	var sourcesSlice []Sources
	for _, s := range sources {
		if s.SourceType == key && s.Active == true {
			sourcesSlice = append(sourcesSlice, s)
		}
	}
	return sourcesSlice
}

func GetActiveSourceByHeadless(sources []Sources, headless bool) []Sources {
	var sourcesSlice []Sources
	for _, s := range sources {
		if s.Headless == headless && s.Active == true {
			sourcesSlice = append(sourcesSlice, s)
		}
	}
	return sourcesSlice
}

func GetSourceByName(sources []Sources, name string) (*Sources, error) {
	for i := 0; i < len(sources); i++ {
		if sources[i].Name == name {
			return &sources[i], nil
		}
	}
	return nil, fmt.Errorf("Unable to find a source with name %s", name)
}

func SourceToNabuPrefix(sources []Sources, useMilled bool) []string {
	jsonld := "summoned"
	if useMilled {
		jsonld = "milled"
	}
	var prefixes []string
	for _, s := range sources {

		switch s.SourceType {

		case "sitemap":
			prefixes = append(prefixes, fmt.Sprintf("%s/%s", jsonld, s.Name))

		case "sitegraph":
			// sitegraph not milled
			prefixes = append(prefixes, fmt.Sprintf("%s/%s", "summoned", s.Name))
		case "googledrive":
			prefixes = append(prefixes, fmt.Sprintf("%s/%s", jsonld, s.Name))
		}
	}
	return prefixes
}
func SourceToNabuProv(sources []Sources) []string {

	var prefixes []string
	for _, s := range sources {

		switch s.SourceType {

		case "sitemap":
			prefixes = append(prefixes, "prov/"+s.Name)

		case "sitegraph":
			prefixes = append(prefixes, "prov/"+s.Name)
		case "googledrive":
			prefixes = append(prefixes, "prov/"+s.Name)
		}
	}
	return prefixes
}

func PruneSources(v1 *viper.Viper, useSources []string) (*viper.Viper, error) {
	var finalSources []Sources
	allSources, err := GetSources(v1)
	if err != nil {
		log.Fatal("error retrieving sources: ", err)
	}
	for _, s := range allSources {
		if contains(useSources, s.Name) {
			s.Active = true // we assume you want to run this, even if disabled, normally
			finalSources = append(finalSources, s)
		}
	}
	if len(finalSources) > 0 {
		v1.Set("sources", finalSources)
		return v1, err
	} else {

		return v1, errors.New("cannot find a source with the name ")
	}

}

// contains checks if a string is present in a slice
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
