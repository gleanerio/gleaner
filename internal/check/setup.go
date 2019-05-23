package check

import (
	"fmt"

	"github.com/minio/minio-go"
)

// MkBuckets sets up the buckets...
func MkBuckets(mc *minio.Client) (bool, error) {

	// loop and make buckets: gleaner, -config -milled -shacl -voc
	// build the buckets...
	err := mc.MakeBucket("gleaner", "us-east-1")
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	fmt.Println("Successfully created mybucket.")

	return true, err
}
