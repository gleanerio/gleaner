package pkg

import (
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/common"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/objects"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/services/bulk"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"

	"github.com/minio/minio-go/v7"
)

func Bulk(v1 *viper.Viper, mc *minio.Client) error {
	//err := bulk.ObjectAssembly(v1, mc)
	err := bulk.BulkAssembly(v1, mc)

	if err != nil {
		log.Error(err)
	}
	return err
}

// used by glcon in gleaner. Need to develop a more common config for the services (aka s3, graph, etc)
// cannot pass a nabu config to the gleaner code to create a minio client, and have it work
func NabuBulk(v1 *viper.Viper) error {
	common.InitLogging()
	mc, err := objects.MinioConnection(v1)
	if err != nil {
		log.Fatal("cannot connect to minio: %s", err)
	}
	return Bulk(v1, mc)
}
