package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"earthcube.org/Project418/gleaner/internal/millers"
	"earthcube.org/Project418/gleaner/internal/summoner"
	"earthcube.org/Project418/gleaner/pkg/utils"
)

var minioVal, portVal, accessVal, secretVal, bucketVal, modeVal, cfgVal string
var sslVal bool

func init() {
	akey := os.Getenv("MINIO_ACCESS_KEY")
	skey := os.Getenv("MINIO_SECRET_KEY")

	flag.StringVar(&minioVal, "address", "localhost", "FQDN for server")
	flag.StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	flag.StringVar(&accessVal, "access", akey, "Access Key ID")
	flag.StringVar(&secretVal, "secret", skey, "Secret access key")
	flag.StringVar(&bucketVal, "bucket", "gleaner", "The configuration bucket")
	flag.StringVar(&cfgVal, "config", "config.json", "Configuration file")
	flag.StringVar(&modeVal, "mode", "cli", "The mode to run in, one of cli or webui")
	flag.BoolVar(&sslVal, "ssl", false, "Use SSL boolean")
}

func main() {
	log.Println("EarthCube Gleaner")

	// Load configurations
	flag.Parse()

	// Look for web..   if seen, go there...
	if strings.EqualFold(modeWal, "webui") {
		webui(cs)
	}

	// Look for cli
	if strings.EqualFold(modeVal, "cli") {

		// Need to look here for cfgVal (file to read) or build from other flag values
		cs := utils.LoadConfigurationS3(minioVal, portVal, accessVal, secretVal, bucketVal, cfgVal, sslVal)

		// if called with a config then validate and load and run
		// if not, walk through a builder (or read all items from command line)
		cli(cs)
	}

}

func webui(cs utils.Config) {

	// If called in "web" mode then
	// Expose a main UI page
	// Expose a service endpoint that validate a utils.Config and loads it
	// Expose a service endpoint that runs gleaner.Summon and gleaner.Mill
	// Expose a status endpoint (easy status from minio calls)

}

func cli(cs utils.Config) {

	// REMOVE once full minio support in...
	// Check for output directory and make it if it doesn't exist
	path := "./deployments/output"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	// REMOVE once full minio support in....
	// Set up an output run directory for each run in output (date)
	t := time.Now()
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	rundir := fmt.Sprintf("%s/%d%d%d_%d%d%d", path, year, month, day, hour, min, sec)
	if _, err := os.Stat(rundir); os.IsNotExist(err) {
		os.Mkdir(rundir, os.ModePerm)
	}

	// Set up our log file for runs...
	// logfile := fmt.Sprintf("%s/logfile.txt", rundir)
	// f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 		log.Fatalf("error opening file: %v", err)
	// 	}
	// 	defer f.Close()
	// 	log.SetOutput(f)

	if cs.Gleaner.Summon {
		summoner.Summoner(cs)
	}

	// could I call these two functions as go func args?
	// NOTE:  summon must run before Mill  !!!!!!!!!
	//	go func() {
	//		time.Sleep(time.Second)
	//		fmt.Println("Go 1")
	//	}()

	if cs.Gleaner.Mill {
		millers.Millers(cs, rundir) // need to remove rundir and then fix the compile
	}
}
