package summoner

import (
	"log"
	"time"

	"earthcube.org/Project418/gleaner/internal/summoner/acquire"
	"github.com/minio/minio-go"
	"github.com/spf13/viper"
)

// Summoner pulls the resources from the data facilities
// func Summoner(mc *minio.Client, cs utils.Config) {
func Summoner(mc *minio.Client, v1 *viper.Viper) {

	log.Printf("Summoner start time: %s \n", time.Now())

	ru := acquire.ResourceURLs(v1, false)
	hru := acquire.ResourceURLs(v1, true)

	if len(ru) > 0 {
		acquire.ResRetrieve(v1, mc, ru)
	}

	if len(hru) > 0 {
		acquire.Headless(mc, hru)
	}

	log.Printf("Summoner end time: %s \n", time.Now())
}
