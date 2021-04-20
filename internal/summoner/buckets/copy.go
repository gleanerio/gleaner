package buckets

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
)

// take the bucket name
// look for bucket.1
// if bucket.1 (empty it)
// copy bucket to bucket.1 now
// empty bucket

func list() {

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	objectCh := minioClient.ListObjects(ctx, "mybucket", minio.ListObjectsOptions{
		Prefix:    "myprefix",
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			fmt.Println(object.Err)
			return
		}
		fmt.Println(object)
	}

}

func copy() {
	// Use-case 1: Simple copy object with no conditions.
	// Source object
	srcOpts := minio.CopySrcOptions{
		Bucket: "my-sourcebucketname",
		Object: "my-sourceobjectname",
	}

	// Destination object
	dstOpts := minio.CopyDestOptions{
		Bucket: "my-bucketname",
		Object: "my-objectname",
	}

	// Copy object call
	uploadInfo, err := minioClient.CopyObject(context.Background(), dst, src)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Successfully copied object:", uploadInfo)
}
