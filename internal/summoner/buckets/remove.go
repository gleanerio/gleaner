package buckets

import (
	"context"
	"fmt"
	"github.com/gleanerio/gleaner/internal/common"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

// empty a bucket (prefix) and remove it

func remove(v1 *viper.Viper) {
	mc := common.MinioConnection(v1)
	//objectsCh := make(chan string)
	objectsCh := make(chan minio.ObjectInfo)

	// Send object names that are needed to be removed to objectsCh
	go func() {
		defer close(objectsCh)
		// List all objects from a bucket-name with a matching prefix.
		opts := minio.ListObjectsOptions{
			Recursive: true,
			Prefix:    "my-prefixname",
		}
		//for object := range mc.ListObjects(context.Background(), "my-bucketname", "my-prefixname", true, nil) {
		for object := range mc.ListObjects(context.Background(), "my-bucketname", opts) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			//objectsCh <- object.Key
			objectsCh <- object
		}

	}()

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}

	for rErr := range mc.RemoveObjects(context.Background(), "my-bucketname", objectsCh, opts) {

		fmt.Println("Error detected during deletion: ", rErr)
	}

	err := mc.RemoveBucket(context.Background(), "mybucket")
	if err != nil {
		fmt.Println(err)
		return
	}

}
