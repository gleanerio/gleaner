package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"earthcube.org/Project418/gleaner/millers"
	"earthcube.org/Project418/gleaner/summoner"
	"earthcube.org/Project418/gleaner/utils"
)

func main() {
	// flags
	// millPtr := flag.Bool("mill", false, "a bool to activate milling")
	// summonPtr := flag.Bool("summon", false, "a bool to request summoning ")
	// cfgFileLoc := flag.String("config", "config.json", "JSON configure file")
	// flag.Parse()

	// load config file
	cs := utils.LoadConfiguration("config.json")

	// Check for output directory and make it if it doesn't exist
	path := "./output" // put in config file????
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

	if cs.Gleaner.Summon {
		summoner.Summoner(cs)
	}

	if cs.Gleaner.Mill {
		millers.Millers(cs, rundir)
	}
}
