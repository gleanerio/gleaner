package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go"
	"github.com/spf13/viper"

	"earthcube.org/Project418/gleaner/internal/check"
	"earthcube.org/Project418/gleaner/internal/common"
	"earthcube.org/Project418/gleaner/internal/millers"
	"earthcube.org/Project418/gleaner/internal/summoner"
)

var viperVal string
var setupVal bool

func init() {
	log.SetFlags(log.Lshortfile)
	// log.SetOutput(ioutil.Discard) // turn off all logging

	// logger approach to buffer in core lib (use for sending logs to s3 in web ui)
	// var (
	// 	buf    bytes.Buffer
	// 	logger = log.New(&buf, "logger: ", log.Lshortfile)
	// )

	flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.StringVar(&viperVal, "cfg", "config", "Configuration file")
}

func main() {
	log.Println("EarthCube Gleaner")
	flag.Parse() // parse any command line flags...

	// Profiling lines (comment out for release builds)
	// defer profile.Start().Stop()                    // cpu
	// defer profile.Start(profile.MemProfile).Stop()  // memory

	v1, err := readConfig(viperVal, map[string]interface{}{
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
	conntest := check.ConnCheck(mc)
	if conntest != nil {
		log.Println("Can not make connection to required object store.  Make sure the minio server is running and accessible")
		os.Exit(1)
	}

	log.Println("Validating access to needed buckets")
	buckets := check.Buckets(mc)
	if buckets != nil {
		log.Printf("%v", buckets)
		os.Exit(1)
	}

	cli(mc, v1)
}

// func cli(mc *minio.Client, cs utils.Config) {
func cli(mc *minio.Client, v1 *viper.Viper) {
	mcfg := v1.GetStringMapString("gleaner")

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
