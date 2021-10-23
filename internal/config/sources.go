package config

import (
	"fmt"
	"github.com/gocarina/gocsv"
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
