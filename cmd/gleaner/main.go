package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"earthcube.org/Project418/gleaner/internal/check"
	"earthcube.org/Project418/gleaner/internal/millers"
	"earthcube.org/Project418/gleaner/internal/summoner"
	"earthcube.org/Project418/gleaner/pkg/utils"

	"github.com/minio/minio-go"
)

var minioVal, portVal, accessVal, secretVal, bucketVal, cfgValObj, cfgValFile string
var sslVal, setupVal bool

func init() {
	log.SetFlags(log.Lshortfile)

	akey := os.Getenv("MINIO_ACCESS_KEY")
	skey := os.Getenv("MINIO_SECRET_KEY")

	flag.StringVar(&minioVal, "address", "localhost", "FQDN for server")
	flag.StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	flag.StringVar(&accessVal, "access", akey, "Access key - read from environment variable if set")
	flag.StringVar(&secretVal, "secret", skey, "Secret key - read from environment variable if set")
	flag.StringVar(&bucketVal, "bucket", "gleaner", "The default bucket namepace")
	flag.StringVar(&cfgValObj, "configobj", "config.json", "Configuration object in object store bucket: [bucket]-config")
	flag.StringVar(&cfgValFile, "configfile", "config.json", "Configuration file")
	flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
	flag.BoolVar(&sslVal, "ssl", false, "Use SSL true/false")
}

func main() {
	log.Println("EarthCube Gleaner")
	flag.Parse() // parse any command line flags...

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	// Create a minio client
	log.Println("Creating needed connection client")
	ep := fmt.Sprintf("%s:%s", minioVal, portVal)
	mc, err := minio.New(ep, accessVal, secretVal, sslVal)
	if err != nil {
		log.Println("Can not create minio connection client")
		os.Exit(1)
	}

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

	cs := utils.Config{}

	if flagset["configfile"] {
		log.Printf("Loading config file: %s \n", cfgValFile)
		cs = utils.ConfigYAML(cfgValFile)
	}

	if flagset["configobj"] {
		log.Printf("Loading config object: %s \n", cfgValObj)
		cs = utils.S3ConfigYAML(mc, "gleaner-config", cfgValObj)
	}

	if !flagset["configfile"] && !flagset["configobj"] {
		fmt.Println("No configuration file or object provided")
		os.Exit(0)
	}

	cli(mc, cs)
}

func cli(mc *minio.Client, cs utils.Config) {
	if cs.Gleaner.Summon {
		summoner.Summoner(mc, cs)
	}

	if cs.Gleaner.Mill {
		millers.Millers(mc, cs) // need to remove rundir and then fix the compile
	}
}
