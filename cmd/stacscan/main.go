package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"time"

	"github.com/gleanerio/gleaner/internal/config"

	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
)

var viperVal, sourceVal, modeVal, logVal string
var setupVal, rudeVal bool

// pass -ldflags "-X main.version=testline"
// go build -ldflags "-X main.version=testline" main.go

var VERSION string

func init() {
	// Output to stdout instead of the default stderr. Can be any io.Writer, see below for File example
	fmt.Println("version: ", VERSION)

	common.InitLogging()

	flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.StringVar(&sourceVal, "source", "", "Override config file source(s) to specify an index target")
	flag.BoolVar(&rudeVal, "rude", false, "Ignore any robots.txt crawl delays or allow / disallow statements")
	flag.StringVar(&viperVal, "cfg", "config", "Configuration file (can be YAML, JSON) Do NOT provide the extension in the command line. -cfg file not -cfg file.yml")
	flag.StringVar(&modeVal, "mode", "full", "Set the mode (full | diff) to index all or just diffs")
	flag.StringVar(&logVal, "log", "warn", "The log level to output (trace | debug | info | warn | error | fatal)")
}

func main() {
	flag.Parse() // parse any command line flags...
	lvl, err := log.ParseLevel(logVal)

	if err != nil {
		log.Fatal("invalid log level:", err.Error())
	}
	log.SetLevel(lvl)

	fmt.Println("STAC Scanner")

	var v1 *viper.Viper

	// Load the config file and set some defaults (config overrides)
	if isFlagPassed("cfg") {
		//v1, err = readConfig(viperVal, map[string]interface{}{})
		v1, err = config.ReadGleanerConfig(filepath.Base(viperVal), filepath.Dir(viperVal))
		if err != nil {
			log.Fatal("error when reading config:", err)
		}
	} else {
		log.Error("Gleaner must be run with a config file: -cfg CONFIGFILE")
		flag.Usage()
		os.Exit(0)
	}

	url := "https://planet.stac.cloud/?t=catalogs"
	timeout := time.Duration(5)
	k := "rda"
	repologger, err := common.LogIssues(v1, k)
	runStats := common.NewRunStats()
	repostats := runStats.Add(k)

	//	acquire.HeadlessNG(v1, mc, hru, db, runStats)
	jsonlds, err := acquire.PageRender(v1, timeout, url, k, repologger, repostats)

	fmt.Println(jsonlds[0])
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
