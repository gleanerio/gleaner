package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"earthcube.org/Project418/gleaner/internal/millers"
	"earthcube.org/Project418/gleaner/internal/summoner"
	"earthcube.org/Project418/gleaner/internal/utils"
)

var minioVal, portVal, accessVal, secretVal, bucketVal, cfgVal string
var sslVal bool

func init() {
	flag.StringVar(&minioVal, "address", "localhost", "FQDN for server")
	flag.StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	flag.StringVar(&accessVal, "access", "AKIAIOSFODNN7EXAMPLE", "Access Key ID")
	flag.StringVar(&secretVal, "secret", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", "Secret access key")
	flag.StringVar(&bucketVal, "bucket", "gleaner", "The configuration bucket")
	flag.StringVar(&cfgVal, "config", "config.json", "Configuration file")
	flag.BoolVar(&sslVal, "ssl", false, "Use SSL boolean")
}

func main() {
	flag.Parse()

	// Load configurations
	cs := utils.LoadConfigurationS3(minioVal, portVal, accessVal, secretVal, bucketVal, cfgVal, sslVal)

	// Check for output directory and make it if it doesn't exist
	path := "./deployments/output"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, os.ModePerm)
	}

	// Set up an output run directory for each run in output (date)
	t := time.Now()
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	rundir := fmt.Sprintf("%s/%d%d%d_%d%d%d", path, year, month, day, hour, min, sec)
	if _, err := os.Stat(rundir); os.IsNotExist(err) {
		os.Mkdir(rundir, os.ModePerm)
	}

	// Set up our log file for runs...
	logfile := fmt.Sprintf("%s/logfile.txt", rundir)
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Println("Gleaner setup complete, starting summoner and millers")

	if cs.Gleaner.Summon {
		summoner.Summoner(cs)
	}

	if cs.Gleaner.Mill {
		millers.Millers(cs, rundir)
	}
}
