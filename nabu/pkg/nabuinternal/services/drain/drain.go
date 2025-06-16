package drain

import (
	"context"

	"github.com/gleanerio/gleaner/nabu/pkg/config"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Objects removes objects from the provided prefixes
func Objects(v1 *viper.Viper, mc *minio.Client) error {
	var err error
	bucketName, err := config.GetBucketName(v1)
	if err != nil {
		return err
	}

	objCfg, _ := config.GetObjectsConfig(v1)
	pa := objCfg.Prefix

	for p := range pa {
		objectsCh := make(chan minio.ObjectInfo)

		// Send object names that are needed to be removed to objectsCh
		go func() {
			defer close(objectsCh)
			// List all objects from a bucket-name with a matching prefix.
			for object := range mc.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{Prefix: pa[p], Recursive: true}) {
				if object.Err != nil {
					log.Println("Error detected during object list in drain: ", object.Err)
					log.Fatalln(object.Err)
				}
				objectsCh <- object
			}
		}()

		opts := minio.RemoveObjectsOptions{
			GovernanceBypass: true,
		}

		for rErr := range mc.RemoveObjects(context.Background(), bucketName, objectsCh, opts) {
			log.Println("Error detected during deletion in drain: ", rErr)
		}

		log.Printf("Remove Bucket: %s  Prefix: %s \n ", bucketName, pa[p])
	}
	return err
}
