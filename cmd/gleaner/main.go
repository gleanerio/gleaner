package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"

	"github.com/gleanerio/gleaner/internal/check"
	"github.com/gleanerio/gleaner/internal/common"
	"github.com/gleanerio/gleaner/internal/millers"
	"github.com/gleanerio/gleaner/internal/objects"
	"github.com/gleanerio/gleaner/internal/organizations"
	"github.com/gleanerio/gleaner/internal/summoner"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
)

var viperVal, sourceVal, modeVal string
var setupVal bool

func init() {
	log.SetFlags(log.Lshortfile)
	// log.SetOutput(ioutil.Discard) // turn off all logging

	flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.StringVar(&sourceVal, "source", "", "Override config file source(s) to specify an index target")
	flag.StringVar(&viperVal, "cfg", "config", "Configuration file (can be YAML, JSON) Do NOT provide the extension in the command line. -cfg file not -cfg file.yml")
	flag.StringVar(&modeVal, "mode", "full", "Set the mode (full | diff) to index all or just diffs")
}

func main() {
	log.Println("EarthCube Gleaner")
	flag.Parse() // parse any command line flags...

	// BEGIN profile section

	// Profiling code (comment out for release builds)
	 //defer profile.Start().Stop()                    // cpu
	 //defer profile.Start(profile.MemProfile).Stop()  // memory

	// Tracing code to use with go tool trace
	 //f, err := os.Create("trace.out")
	 //if err != nil {
		 //panic(err)
	 //}
	 //defer f.Close()

	 //err = trace.Start(f)
	 //if err != nil {
		 //panic(err)
	 //}
	 //defer trace.Stop()

	// END profile section

	var v1 *viper.Viper
	var err error

	// Load the config file and set some defaults (config overrides)
	if isFlagPassed("cfg") {
		v1, err = readConfig(viperVal, map[string]interface{}{
			"minio": map[string]string{
				"address":   "localhost",
				"port":      "9000",
				"accesskey": "",
				"secretkey": "",
			},
		})
		if err != nil {
			log.Printf("error when reading config: %v", err)
			os.Exit(1)
		}
	} else {
		log.Println("Gleaner must be run with a config file: -cfg CONFIGFILE")
		flag.Usage()
		os.Exit(0)
	}

	// read config file for minio info (need the bucket to check existence )
	miniocfg := v1.GetStringMapString("minio")
	bucketName := miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file

	// Remove all source EXCEPT the one in the source command lind
	if isFlagPassed("source") {
		tmp := []objects.Sources{} // tmp slice to hold our desired source

		var domains []objects.Sources
		err := v1.UnmarshalKey("sources", &domains)
		if err != nil {
			log.Println(err)
		}

		for _, k := range domains {
			if sourceVal == k.Name {
				tmp = append(tmp, k)
			}
		}

		if len(tmp) == 0 {
			log.Println("CAUTION:  no sources, did your -source VALUE match a sources.name VALUE in your config file?")
			os.Exit(0)
		}

		configMap := v1.AllSettings()
		delete(configMap, "sources")
		v1.Set("sources", tmp)
	}

	// Parse a new mode entry from command line if present
	if isFlagPassed("mode") {
		m := v1.GetStringMap("summoner")
		m["mode"] = modeVal
		v1.Set("summoner", m)
	}

	// Set up the minio connector
	mc := common.MinioConnection(v1)

	// If requested, set up the buckets
	if setupVal {
		log.Println("Setting up buckets")
		err := check.MakeBuckets(mc, bucketName)
		if err != nil {
			log.Println("Error making buckets for setup call")
			os.Exit(1)
		}

		log.Println("Buckets generated.  Object store should be ready for runs")
		os.Exit(0)
	}

	// Validate Minio is up  TODO:  validate all expected containers are up
	log.Println("Validating access to object store")
	err = check.ConnCheck(mc)
	if err != nil {
		log.Printf("Connection issue, make sure the minio server is running and accessible. %s ", err)
		os.Exit(1)
	}

	err = check.Buckets(mc, bucketName)
	if err != nil {
		log.Printf("Can not find bucket. %s ", err)
		os.Exit(1)
	}

	cli(mc, v1)
}

// func cli(mc *minio.Client, cs utils.Config) {
func cli(mc *minio.Client, v1 *viper.Viper) {
	mcfg := v1.GetStringMapString("gleaner")

	// Build the org graph
	// err := organizations.BuildGraphMem(mc, v1) // parquet testing
	err := organizations.BuildGraph(mc, v1)
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

func readConfig(filename string, defaults map[string]interface{}) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
	v.SetConfigName(filename)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
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

// func to support remove elements form the source slice
func remove(s []objects.Sources, i int) []objects.Sources {
	fmt.Println("removing")

	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
