package objects

import (
	"context"
	"github.com/gleanerio/gleaner/nabu/pkg/config"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"

	"github.com/minio/minio-go/v7"
	"github.com/spf13/viper"
)

// ObjectAssembly collects the objects from a bucket to load
func ObjectAssembly(v1 *viper.Viper, mc *minio.Client) error {
	objs, err := config.GetObjectsConfig(v1)
	//spql, err := config.GetSparqlConfig(v1)
	ep := v1.GetString("flags.endpoint")
	spql, err := config.GetEndpoint(v1, ep, "bulk")
	if err != nil {
		log.Error(err)
	}

	var pa = objs.Prefix

	//if strings.Contains(strings.Join(pa, ","), s) {
	//	fmt.Println(s, "is in the array")
	//}

	for p := range pa {
		oa := []string{}

		// NEW
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		bucketName, _ := config.GetBucketName(v1)
		objectCh := mc.ListObjects(ctx, bucketName, minio.ListObjectsOptions{Prefix: pa[p], Recursive: true})

		for object := range objectCh {
			if object.Err != nil {
				log.Error(object.Err)
				return object.Err
			}
			// fmt.Println(object)
			oa = append(oa, object.Key)
		}

		log.Infof("%s:%s object count: %d\n", bucketName, pa[p], len(oa))
		bar := progressbar.Default(int64(len(oa)))
		for item := range oa {
			_, err := PipeLoad(v1, mc, bucketName, oa[item], spql.URL)
			if err != nil {
				log.Error(err)
			}
			bar.Add(1)
			// log.Println(string(s)) // get "s" on pipeload and send to a log file
		}
	}

	return err
}
