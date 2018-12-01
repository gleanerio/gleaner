package summoner

import (
	"log"
	"time"

	"earthcube.org/Project418/gleaner/internal/summoner/acquire"
	"earthcube.org/Project418/gleaner/internal/utils"
)

// Summoner pulls the resources from the data facilities
func Summoner(cs utils.Config) {
	log.Printf("Summoner start time: %s \n", time.Now()) // Log start time

	domains, headlessdomains, err := acquire.DomainListJSON(cs)
	if err != nil {
		log.Printf("Error reading list of domains %v\n", err)
	}

	log.Printf("Domains: %v \n", domains)
	log.Printf("Headless domains: %v \n", headlessdomains)

	// TODO  the following two functions could be done concurrently
	ru := acquire.ResourceURLs(domains, cs) // map by domain name and []string of landing page URLs
	if len(ru) > 0 {
		acquire.ResRetrieve(ru, cs)
	}

	hru := acquire.ResourceURLs(headlessdomains, cs) // map by domain name and []string of landing page URLs
	if len(hru) > 0 {
		acquire.Headless(hru, cs)
	}

	log.Printf("Summoner end time: %s \n", time.Now()) // Log end time
}
