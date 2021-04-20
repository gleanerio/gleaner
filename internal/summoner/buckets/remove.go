package buckets

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
)

// empty a bucket (prefix) and remove it

func remove() {

	objectsCh := make(chan string)

	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		for object := range minioClient.ListObjects(context.Background(), "my-bucketname", "my-prefixname", true, nil) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object.Key
		}

	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	for rErr := range minioClient.RemoveObjects(context.Background(), "my-bucketname", objectsCh, opts) {
		fmt.Println("Error detected during deletion: ", rErr)
	}

	err = minioClient.RemoveBucket(context.Background(), "mybucket")
	if err != nil {
		fmt.Println(err)
		return
	}

}
