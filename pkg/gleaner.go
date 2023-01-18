package pkg

import (
	"github.com/gleanerio/gleaner/internal/check"
	"github.com/gleanerio/gleaner/internal/millers"
	"github.com/gleanerio/gleaner/internal/organizations"
	"github.com/gleanerio/gleaner/internal/summoner"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
	"os"
	//"os"
)

func Cli(mc *minio.Client, v1 *viper.Viper, db *bolt.DB) error {

	mcfg := v1.GetStringMapString("gleaner")
	//mcfg := v1.Sub("gleaner") /// with overrides from batch ends up being nil
	err := check.PreflightChecks(mc, v1)
	if err != nil {
		log.Fatal("Failed Preflight connection check to minio. Check configuration", err)
		os.Exit(66)
	}
	// Build the org graph
	// err := organizations.BuildGraphMem(mc, v1) // parfquet testing
	err = organizations.BuildGraph(mc, v1)
	if err != nil {
		log.Error(err)
	}

	// If configured, summon sources
	if mcfg["summon"] == "true" {
		// Index the sitegraphs first, if any
		fn, err := acquire.GetGraph(mc, v1)
		if err != nil {
			log.Error(err)
		}
		log.Info(fn)
		// summon sitemaps
		summoner.Summoner(mc, v1, db)
		acquire.GetFromGDrive(mc, v1)
	}

	// if configured, process summoned sources fronm JSON-LD to RDF (nq)
	if mcfg["mill"] == "true" {
		millers.Millers(mc, v1) // need to remove rundir and then fix the compile
	}
	return err
}
