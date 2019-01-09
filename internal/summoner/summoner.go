package summoner

import (
	"log"
	"time"

	"earthcube.org/Project418/gleaner/internal/summoner/acquire"
	"earthcube.org/Project418/gleaner/pkg/utils"
)

// Summoner pulls the resources from the data facilities
func Summoner(cs utils.Config) {
	log.Printf("Summoner start time: %s \n", time.Now())

	domains, headlessdomains, err := acquire.DomainListJSON(cs)
	if err != nil {
		log.Printf("Error reading list of domains %v\n", err)
	}

	log.Printf("Domains: %v \n", domains)
	log.Printf("Headless domains: %v \n", headlessdomains)

	// NOTE: Following two functions could be modified to run concurrently (both between each other and internal)
	ru := acquire.ResourceURLs(domains, cs)
	if len(ru) > 0 {
		acquire.ResRetrieve(ru, cs)
	}

	hru := acquire.ResourceURLs(headlessdomains, cs)
	if len(hru) > 0 {
		acquire.Headless(hru, cs)
	}

	log.Printf("Summoner end time: %s \n", time.Now())
}
