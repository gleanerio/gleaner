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
var sslVal, checkVal bool

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
	flag.BoolVar(&checkVal, "check", false, "Run Gleaner configuration check and exit")
	flag.BoolVar(&sslVal, "ssl", false, "Use SSL true/false")
}

func main() {
	log.Println("EarthCube Gleaner")
	flag.Parse() // parse any command line flags...

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	// Get a connection to our minio
	ep := fmt.Sprintf("%s:%s", minioVal, portVal)
	mc, err := minio.New(ep, accessVal, secretVal, sslVal)
	if err != nil {
		log.Println("Can not make connection to required object store.")
		log.Fatalln(err)
	}

	// Look for checksetup and check connections and buckets...
	if checkVal {
		check.GleanerCheck(mc)
		os.Exit(0)
	}

	// if setupVal {
	// set up the buckets
	// os.Exit(0)
	// }

	cs := utils.Config{}

	if flagset["configfile"] {
		log.Printf("Loading config file: %s \n", cfgValFile)
		// cs = utils.LoadConfiguration(cfgValFile)
		cs = utils.ConfigYAML(cfgValFile)
	}

	if flagset["configobj"] {
		log.Printf("Loading config object: %s \n", cfgValObj)
		cs = utils.LoadConfigurationS3(mc, "gleaner-config", cfgValObj)
	}

	if !flagset["configfile"] && !flagset["configobj"] {
		fmt.Println("No configuration file provided")
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
