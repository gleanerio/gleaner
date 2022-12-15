package summoner

import (
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/gleanerio/gleaner/internal/summoner/acquire"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
)

// Summoner pulls the resources from the data facilities
// func Summoner(mc *minio.Client, cs utils.Config) {
func Summoner(mc *minio.Client, v1 *viper.Viper, db *bolt.DB) {
	st := time.Now()
	log.Info("Summoner start time:", st) // Log the time at start for the record
	runStats := common.NewRunStats()

	// Retrieve API urls
	apiSources, err := api.RetrieveAPIEndpoints(vi)
	if err != nil {
		log.Error("Error getting API endpoint sources:", err)
	} else if len(api) > 0 {
		api.RetrieveAPIData(apiSources, mc, db, runStats)
	}

	// Get a list of resource URLs that do and don't require headless processing
	ru, err := acquire.ResourceURLs(v1, mc, false, db)
	if err != nil {
		log.Info("Error getting urls that do not require headless processing:", err)
	}
	// just report the error, and then run gathered urls
	if len(ru) > 0 {
		acquire.ResRetrieve(v1, mc, ru, db, runStats) // TODO  These can be go funcs that run all at the same time..
	}

	hru, err := acquire.ResourceURLs(v1, mc, true, db)
	if err != nil {
		log.Info("Error getting urls that require headless processing:", err)
	}
	// just report the error, and then run gathered urls
	if len(hru) > 0 {
		acquire.HeadlessNG(v1, mc, hru, db, runStats)
	}

	// Time report
	et := time.Now()
	diff := et.Sub(st)
	log.Info("Summoner end time:", et)
	log.Info("Summoner run time:", diff.Minutes())
	fmt.Print(runStats.Output())

	// What do I need to the "run" prov
	// the URLs indexed  []string
	// the graph generated?  "version" the graph by the build date
	// pass ru, hru, and v1 to a run prov function.
	//	RunFeed(v1, mc, et, ru, hru)  // DEV:   hook for building feed  (best place for it?)

}
