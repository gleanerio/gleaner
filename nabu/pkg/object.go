package pkg

import (
	"fmt"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/common"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/objects"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/services/bulk"
	//	"github.com/gleanerio/gleaner/nabu/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

func Object(v1 *viper.Viper, mc *minio.Client, bucket string, object string) error {
	fmt.Println("Load graph object to triplestore")
	//	spql, _ := config.GetSparqlConfig(v1)
	//	if bucket == "" {
	//		bucket, _ = config.GetBucketName(v1)
	//	}
	//	s, err := objects.PipeLoad(v1, mc, bucket, object, spql.Endpoint)
	//	if err != nil {
	//		log.Error(err)
	//	}

	s, err := bulk.BulkLoad(v1, mc, bucket, object)
	if err != nil {
		log.Println(err)
	}

	log.Trace(string(s))
	return err
}

// used by glcon in gleaner. Need to develop a more common config for the services (aka s3, graph, etc)
// cannot pass a nabu config to the gleaner code to create a minio client, and have it work
func NabuObject(v1 *viper.Viper, bucket string, object string) error {
	common.InitLogging()
	mc, err := objects.MinioConnection(v1)
	if err != nil {
		log.Fatal("cannot connect to minio: %s", err)
	}
	return Object(v1, mc, bucket, object)
}
