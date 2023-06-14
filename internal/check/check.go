package check

import (
	"context"
	"fmt"
	"github.com/gleanerio/gleaner/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

// ConnCheck check the connections with a list buckets call
func ConnCheck(mc *minio.Client) error {

	buckets, err := mc.ListBuckets(context.Background())
	log.Trace(buckets)

	return err
}

func isExists(value string, data []minio.BucketInfo) (exists bool) {

	for _, search := range data {
		if search.Name == value {
			return true
		}
	}
	return false
}

// Buckets checks the setup
func Buckets(mc *minio.Client, bucket string) error {
	var err error

	// for i := range bl {
	//found, err := mc.BucketExists(context.Background(), bucket) // returns a redirect if region is not correct
	buckets, err := mc.ListBuckets(context.Background())
	found := isExists(bucket, buckets)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("unable to locate required bucket:  %s, did you run gleaner with -setup the first to set up buckets?", bucket)
	}
	if found {
		log.Debug("Validated access to object store:", bucket)
	}
	// }

	return err
}

// MakeBuckets checks the setup
func MakeBuckets(mc *minio.Client, bucket string) error {
	var err error

	// for i := range bl {
	found, err := mc.BucketExists(context.Background(), bucket)
	if err != nil {
		log.Debug("Existing bucket", bucket, "check:", err)
	}
	if found {
		log.Debug("Gleaner Bucket", bucket, "found.")
	} else {
		log.Debug("Gleaner Bucket", bucket, "not found, generating")
		err = mc.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: "us-east-1"}) // location is kinda meaningless here
		if err != nil {
			log.Debug("Make bucket:", err)
		}
	}
	// }

	return err
}

/*
*
Setup Gleaner buckets
*/
func Setup(mc *minio.Client, v1 *viper.Viper) error {
	ms := v1.Sub("minio")
	m1, err := config.ReadMinioConfig(ms)
	if err != nil {
		log.Error("Error reading gleaner config", err)
		return err
	}
	// Validate Minio is up  TODO:  validate all expected containers are up
	log.Info("Validating access to object store")
	err = ConnCheck(mc)
	if err != nil {
		log.Error("Connection issue, make sure the minio server is running and accessible.", err)
		return err
	}
	// If requested, set up the buckets
	log.Info("Setting up buckets")
	err = MakeBuckets(mc, m1.Bucket)
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
*/
func PreflightChecks(mc *minio.Client, v1 *viper.Viper) error {
	// Validate Minio access
	bucketName, err := config.GetBucketName(v1)

	if err != nil {
		log.Error("missing bucket name.", err)
		return err
	}
	err = ConnCheck(mc)
	if err != nil {
		log.Error("Connection issue, make sure the minio server is running and accessible.", err)
		return err
	}
	//Check our bucket is ready
	err = Buckets(mc, bucketName)
	if err != nil {
		log.Error("Can not find bucket.", err)
		return err
	}
	return nil
}
