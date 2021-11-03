package check

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
)

// ConnCheck check the connections with a list buckets call
func ConnCheck(mc *minio.Client) error {
	_, err := mc.ListBuckets(context.Background())
	return err
}

// Buckets checks the setup
func Buckets(mc *minio.Client, bucket string) error {
	var err error

	// for i := range bl {
	found, err := mc.BucketExists(context.Background(), bucket)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("unable to locate required bucket:  %s, did you run gleaner with -setup the first to set up buckets?", bucket)
	}
	if found {
		log.Printf("Validated access to object store: %s\n", bucket)
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
		log.Printf("Existing bucket %s check: %v\n", bucket, err)
	}
	if found {
		log.Printf("Gleaner Bucket %s found.\n", bucket)
	} else {
		log.Printf("Gleaner Bucket %s not found, generating\n", bucket)
		err = mc.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: "us-east-1"}) // location is kinda meaningless here
		if err != nil {
			log.Printf("Make bucket: %v\n", err)
		}
	}
	// }

	return err
}
