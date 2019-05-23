package summoner

import (
	"log"
	"time"

	"earthcube.org/Project418/gleaner/internal/summoner/acquire"
	"earthcube.org/Project418/gleaner/pkg/utils"
	"github.com/minio/minio-go"
)

// Summoner pulls the resources from the data facilities
func Summoner(mc *minio.Client, cs utils.Config) {
	log.Printf("Summoner start time: %s \n", time.Now())

	domains, headlessdomains, err := acquire.DomainListJSON(cs)
	if err != nil {
		log.Printf("Error reading list of domains %v\n", err)
	}

	log.Printf("Domains: %v \n", domains)
	log.Printf("Headless domains: %v \n", headlessdomains)

	ru := acquire.ResourceURLs(domains, cs)
	if len(ru) > 0 {
		acquire.ResRetrieve(mc, ru, cs)
	}

	hru := acquire.ResourceURLs(headlessdomains, cs) // TODO..  pass mc and get this working again
	if len(hru) > 0 {
		acquire.Headless(mc, hru, cs)
	}

	log.Printf("Summoner end time: %s \n", time.Now())
}
