package pkg

import (
	"github.com/gleanerio/gleaner/internal/check"
	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/gleanerio/gleaner/internal/millers"
	"github.com/gleanerio/gleaner/internal/organizations"
	"github.com/gleanerio/gleaner/internal/summoner"
	"github.com/gleanerio/gleaner/internal/summoner/acquire"
	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
	"log"
	//"os"
)

func Cli(mc *minio.Client, v1 *viper.Viper, db *bolt.DB) error {

	mcfg := v1.GetStringMapString("gleaner")

	// Build the org graph
	// err := organizations.BuildGraphMem(mc, v1) // parfquet testing
	err := organizations.BuildGraph(mc, v1)
	if err != nil {
		log.Print(err)
	}

	// If configured, summon sources
	if mcfg["summon"] == "true" {
		// Index the sitegraphs first, if any
		fn, err := acquire.GetGraph(mc, v1)
		if err != nil {
			log.Print(err)
		}
		log.Println(fn)
		// summon sitemaps
		summoner.Summoner(mc, v1, db)
	}

	// if configured, process summoned sources fronm JSON-LD to RDF (nq)
	if mcfg["mill"] == "true" {
		millers.Millers(mc, v1) // need to remove rundir and then fix the compile
	}
	return err
}

/**
Setup Gleaner buckets

*/
func Setup(mc *minio.Client, v1 *viper.Viper) error {
	ms := v1.Sub("minio")
	m1, err := configTypes.ReadMinioConfig(ms)
	if err != nil {
		log.Printf("Error reading gleaner config %s ", err)
		return err
	}
	// Validate Minio is up  TODO:  validate all expected containers are up
	log.Println("Validating access to object store")
	err = check.ConnCheck(mc)
	if err != nil {
		log.Printf("Connection issue, make sure the minio server is running and accessible. %s ", err)
		return err
	}
	// If requested, set up the buckets
	log.Println("Setting up buckets")
	err = check.MakeBuckets(mc, m1.Bucket)
	if err != nil {
		log.Println("Error making buckets for Setup call")
		return err
	}

	err = PreflightChecks(mc, v1) // postsetup test ;)
	if err != nil {
		return err
	}
	log.Println("Buckets generated.  Object store should be ready for runs")
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
		log.Printf("Connection issue, make sure the minio server is running and accessible. %s ", err)
		return err
	}

	// Check our bucket is ready
	err = check.Buckets(mc, bucketName)
	if err != nil {
		log.Printf("Can not find bucket. %s ", err)
		return err
	}
	return nil
}
