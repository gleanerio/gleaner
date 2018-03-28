package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"earthcube.org/Project418/gleaner/summoner/acquire"
	"earthcube.org/Project418/gleaner/summoner/utils"
)

func main() {
	// Set up our log file for runs...
	f, err := os.OpenFile("logfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Printf("Start time: %s \n", time.Now()) // Log start time

	cfgFileLoc := flag.String("config", "config.json", "JSON configure file")
	flag.Parse()

	cs := loadConfiguration(cfgFileLoc)

	domains, headlessdomains, err := acquire.DomainListJSON(cs.Source)
	if err != nil {
		log.Printf("Error reading list of domains %v\n", err)
	}

	fmt.Println(domains)
	fmt.Println(headlessdomains)

	// TODO  the following two loops could be donen concurrently
	ru := acquire.ResourceURLsJSON(domains) // map by domain name and []string of landing page URLs
	if len(ru) > 0 {
		acquire.ResRetrieve(ru, cs)
	}

	hru := acquire.ResourceURLsJSON(headlessdomains) // map by domain name and []string of landing page URLs
	if len(hru) > 0 {
		acquire.Headless(hru, cs)
	}

	log.Printf("End time: %s \n", time.Now()) // Log end time
}

func loadConfiguration(file *string) utils.Config {
	var config utils.Config
	configFile, err := os.Open(*file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}
