package objects

import (
	"context"
	"github.com/gleanerio/gleaner/nabu/pkg/config"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// GetObjects is the generic object collection function
func GetObjects(v1 *viper.Viper, mc *minio.Client) ([]string, error) {
	oa := []string{}

	objs, err := config.GetObjectsConfig(v1)
	var pa = objs.Prefix

	log.Println(pa)

	for p := range pa {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		bucketName, _ := config.GetBucketName(v1)
		objectCh := mc.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: pa[p], Recursive: true})

		for object := range objectCh {
			if object.Err != nil {
				log.Println(object.Err)
				return oa, object.Err
			}
			// fmt.Println(object)
			oa = append(oa, object.Key)
		}
		log.Printf("%s:%s object count: %d\n", bucketName, pa[p], len(oa))
	}

	return oa, err
}
