package summoner

import (
	"log"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// Summoner pulls the resources from the data facilities
// func Summoner(mc *minio.Client, cs utils.Config) {
func Summoner(mc *minio.Client, v1 *viper.Viper, db *bolt.DB) {
	st := time.Now()
	log.Printf("Summoner start time: %s \n", st) // Log the time at start for the record

	// Get a list of resource URLs that do and don't required headless processing
	ru := acquire.ResourceURLs(v1, mc, false, db)
	hru := acquire.ResourceURLs(v1, mc, true, db)

	// TODO  These can be go funcs that run all at the same time..
	if len(ru) > 0 {
		acquire.ResRetrieve(v1, mc, ru, db)
	}

	if len(hru) > 0 {
		acquire.HeadlessNG(v1, mc, hru, db)
	}

	// Time report
	et := time.Now()
	diff := et.Sub(st)
	log.Printf("Summoner end time: %s \n", et)
	log.Printf("Summoner run time: %f \n", diff.Minutes())

}
