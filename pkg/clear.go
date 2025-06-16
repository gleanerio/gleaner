package pkg

import (
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/common"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/graph"
	"github.com/gleanerio/gleaner/nabu/pkg/nabuinternal/objects"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Clear(v1 *viper.Viper, mc *minio.Client) error {
	log.Info("Nabu started with mode: clear")

	d := v1.GetBool("flags.dangerous")

	if d {
		log.Println("dangerous mode is enabled")
		_, err := graph.Clear(v1)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		log.Println("dangerous mode must be set to true to run this")
		return nil
	}

	return nil
}

// NabuClear used by glcon in gleaner. Need to develop a more common config for the services (aka s3, graph, etc.)
// cannot pass a nabu config to the gleaner code to create a minio client, and have it work
func NabuClear(v1 *viper.Viper) error {
	common.InitLogging()
	mc, err := objects.MinioConnection(v1)
	if err != nil {
		log.Fatal("cannot connect to minio: %s", err)
	}
	return Prefix(v1, mc)
}
