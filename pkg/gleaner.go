package pkg

import (
	"github.com/gleanerio/gleaner/internal/check"
	configTypes "github.com/gleanerio/gleaner/internal/config"
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

/*
*
Setup Gleaner buckets
*/
func Setup(mc *minio.Client, v1 *viper.Viper) error {
	ms := v1.Sub("minio")
	m1, err := configTypes.ReadMinioConfig(ms)
	if err != nil {
		log.Error("Error reading gleaner config", err)
		return err
	}
	// Validate Minio is up  TODO:  validate all expected containers are up
	log.Info("Validating access to object store")
	err = check.ConnCheck(mc)
	if err != nil {
		log.Error("Connection issue, make sure the minio server is running and accessible.", err)
		return err
	}
	// If requested, set up the buckets
	log.Info("Setting up buckets")
	err = check.MakeBuckets(mc, m1.Bucket)
	if err != nil {
		log.Error("Error making buckets for Setup call", err)
		return err
	}

	err = PreflightChecks(mc, v1) // postsetup test ;)
	if err != nil {
		return err
	}
	log.Info("Buckets generated.  Object store should be ready for runs")
	return nil

}

/*
Check to see we can connect to s3 instance, and that buckets exist
Might also be used to flight check bolt database, and if containers are up
*/
func PreflightChecks(mc *minio.Client, v1 *viper.Viper) error {
	// Validate Minio access
	bucketName, err := configTypes.GetBucketName(v1)

	err = check.ConnCheck(mc)
	if err != nil {
		log.Error("Connection issue, make sure the minio server is running and accessible.", err)
		return err
	}

	// Check our bucket is ready
	err = check.Buckets(mc, bucketName)
	if err != nil {
		log.Error("Can not find bucket.", err)
		return err
	}
	return nil
}
