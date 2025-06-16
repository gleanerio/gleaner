package objects

import (
	"context"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Remove is the generic object collection function
func Remove(v1 *viper.Viper, mc *minio.Client, bucket, object string) error {
	opts := minio.RemoveObjectOptions{
		GovernanceBypass: true,
		//VersionID:        "myversionid",
	}

	err := mc.RemoveObject(context.Background(), bucket, object, opts)
	if err != nil {
		log.Println(err)
		return err
	}

	return err
}
