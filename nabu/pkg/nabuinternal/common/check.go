package common

import (
	"context"
	"github.com/minio/minio-go/v7"
)

// ConnCheck check the connections with a list buckets call
func ConnCheck(mc *minio.Client) error {
	_, err := mc.ListBuckets(context.Background())
	return err
}
