package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go"
	"github.com/spf13/viper"

	"github.com/earthcubearchitecture-project418/gleaner/internal/check"
	"github.com/earthcubearchitecture-project418/gleaner/internal/common"
	"github.com/earthcubearchitecture-project418/gleaner/internal/millers"
	"github.com/earthcubearchitecture-project418/gleaner/internal/organizations"
	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner"
	"github.com/earthcubearchitecture-project418/gleaner/internal/summoner/acquire"
)

var viperVal, sourceVal string
var setupVal bool

func init() {
	log.SetFlags(log.Lshortfile)
	// log.SetOutput(ioutil.Discard) // turn off all logging

	flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.StringVar(&sourceVal, "source", "", "Override config file source")
	flag.StringVar(&viperVal, "cfg", "config", "Configuration file")
}

func main() {
	log.Println("EarthCube Gleaner")
	flag.Parse() // parse any command line flags...

	// BEGIN profile section
	// Profiling code (comment out for release builds)
	// defer profile.Start().Stop()                    // cpu
	// defer profile.Start(profile.MemProfile).Stop()  // memory

	// Tracing code to use with go tool trace
	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()

	// END profile section

	var v1 *viper.Viper
	var err error

	// Load the config file and set some defaults (config overrides)
	if isFlagPassed("cfg") {
		v1, err = readConfig(viperVal, map[string]interface{}{
			"sqlfile": "",
			"bucket":  "",
			"minio": map[string]string{
				"address":   "localhost",
				"port":      "9000",
				"accesskey": "",
				"secretkey": "",
			},
		})
		if err != nil {
			panic(fmt.Errorf("error when reading config: %v", err))
		}
	}

	if isFlagPassed("source") {
		// -source '{"Name":"demo1", "URL":"http://foo.org/sitemap.xml", "Headless":true}'

		configMap := v1.AllSettings()
		delete(configMap, "sources")

		//log.Println(sourceVal)
		ns := acquire.Sources{}
		json.Unmarshal([]byte(sourceVal), &ns)
		//log.Println(ns)

		sa := []acquire.Sources{}
		//s := acquire.Sources{Name: "dmeoname", URL: "https://foo.org/sitemap.xml", Headless: true}
		//sa = append(sa, s)
		sa = append(sa, ns)
		v1.Set("sources", sa)
	}

	//var domains []acquire.Sources
	//err = v1.UnmarshalKey("sources", &domains)
	//if err != nil {
	//log.Println(err)
	//}
	//log.Println(domains)
	//os.Exit(0)

	// Set up the minio connector
	mc := common.MinioConnection(v1)

	// If requested, set up the buckets
	if setupVal {
		log.Println("Setting up buckets")
		err := check.MakeBuckets(mc)
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

	err = check.Buckets(mc)
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
	organizations.BuildGraph(mc, v1)

	os.Exit(0)

	if mcfg["summon"] == "true" {
		summoner.Summoner(mc, v1)
	}

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
