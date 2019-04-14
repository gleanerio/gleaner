package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"earthcube.org/Project418/gleaner/internal/millers"
	"earthcube.org/Project418/gleaner/internal/summoner"
	"earthcube.org/Project418/gleaner/pkg/utils"
	"github.com/minio/minio-go"
)

// TODO
// make a struct for cs and mc and pass it around to things as a simple pointer

var minioVal, portVal, accessVal, secretVal, bucketVal, modeVal, cfgVal string
var sslVal bool

func init() {
	log.SetFlags(log.Lshortfile)

	akey := os.Getenv("MINIO_ACCESS_KEY")
	skey := os.Getenv("MINIO_SECRET_KEY")

	flag.StringVar(&minioVal, "address", "localhost", "FQDN for server")
	flag.StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	flag.StringVar(&accessVal, "access", akey, "Access Key ID")
	flag.StringVar(&secretVal, "secret", skey, "Secret access key")
	flag.StringVar(&bucketVal, "bucket", "gleaner", "The configuration bucket")
	flag.StringVar(&cfgVal, "config", "", "Configuration file")
	flag.StringVar(&modeVal, "mode", "cli", "The mode to run in, one of cli or webui")
	flag.BoolVar(&sslVal, "ssl", false, "Use SSL boolean")
}

func main() {
	log.Println("EarthCube Gleaner")

	// Load configurations
	flag.Parse()

	if !strings.EqualFold(modeVal, "cli") && !strings.EqualFold(modeVal, "webui") {
		fmt.Println("Mode needs to be set to one of cli or webui")
		log.Fatal("Mode not set")
	}

	// Get a connection to our minio
	ep := fmt.Sprintf("%s:%s", minioVal, portVal)
	mc, err := minio.New(ep, accessVal, secretVal, sslVal)
	if err != nil {
		log.Fatalln(err)
	}

	// Look for web..   if seen, go there...
	if strings.EqualFold(modeVal, "webui") {
		// web ui will need to know the S3 info....
		webui()
	}

	// Look for cli
	if strings.EqualFold(modeVal, "cli") {
		cs := utils.Config{}

		// Config file is either provided at command line, or we look for it in S3/Minio
		if !strings.EqualFold(cfgVal, "") { // this is stupid..  do this better
			cs = utils.LoadConfiguration(cfgVal)
		} else {
			// cs = utils.LoadConfigurationS3(minioVal, portVal, accessVal, secretVal, bucketVal, cfgVal, sslVal)
			cs = utils.LoadConfigurationS3(mc, "gleaner-config", "config.json")
		}
		cli(mc, cs)
	}
}

func webui() {
	fmt.Println("In development...")
}

func cli(mc *minio.Client, cs utils.Config) {
	// REMOVE once full minio support in...
	// Check for output directory and make it if it doesn't exist
	// path := "./deployments/output"
	// if _, err := os.Stat(path); os.IsNotExist(err) {
	// 	os.Mkdir(path, os.ModePerm)
	// }

	// REMOVE once full minio support in....
	// Set up an output run directory for each run in output (date)
	// t := time.Now()
	// year, month, day := t.Date()
	// hour, min, sec := t.Clock()
	// rundir := fmt.Sprintf("%s/%d%d%d_%d%d%d", path, year, month, day, hour, min, sec)
	// if _, err := os.Stat(rundir); os.IsNotExist(err) {
	// 	os.Mkdir(rundir, os.ModePerm)
	// }

	// Set up our log file for runs...
	// logfile := fmt.Sprintf("%s/logfile.txt", rundir)
	// f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 		log.Fatalf("error opening file: %v", err)
	// 	}
	// 	defer f.Close()
	// 	log.SetOutput(f)

	// could I call these two functions as go func args?
	// https://abronan.com/introduction-to-goroutines-and-go-channels/
	// NOTE:  summon must run before mill

	if cs.Gleaner.Summon {
		summoner.Summoner(mc, cs)
	}

	if cs.Gleaner.Mill {
		// millers.Millers(cs, rundir) // need to remove rundir and then fix the compile
		millers.Millers(mc, cs) // need to remove rundir and then fix the compile
	}
}
