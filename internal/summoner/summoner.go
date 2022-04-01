package summoner

import (
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// Summoner pulls the resources from the data facilities
// func Summoner(mc *minio.Client, cs utils.Config) {
func Summoner(mc *minio.Client, v1 *viper.Viper) {
	st := time.Now()
	log.Printf("Summoner start time: %s \n", st) // Log the time at start for the record

	// Get a list of resource URLs that do and don't require headless processing
	ru, err := acquire.ResourceURLs(v1, mc, false)
	if err != nil {
		log.Printf("Error getting urls that do not require headless processing: %s", err)
	} else if len(ru) > 0 {
		acquire.ResRetrieve(v1, mc, ru) // TODO  These can be go funcs that run all at the same time..
	}

	hru, err := acquire.ResourceURLs(v1, mc, true)
	if err != nil {
		log.Printf("Error getting urls that require headless processing: %s", err)
	} else if len(hru) > 0 {
		acquire.HeadlessNG(v1, mc, hru)
	}

	// Time report
	et := time.Now()
	diff := et.Sub(st)
	log.Printf("Summoner end time: %s \n", et)
	log.Printf("Summoner run time: %f \n", diff.Minutes())

	// What do I need to the "run" prov
	// the URLs indexed  []string
	// the graph generated?  "version" the graph by the build date
	// pass ru, hru, and v1 to a run prov function.
	//	RunFeed(v1, mc, et, ru, hru)  // DEV:   hook for building feed  (best place for it?)

}
