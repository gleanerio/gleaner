package objects

import (
	"context"
	"github.com/minio/minio-go/v7"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Copy is the generic object collection function
func Copy(v1 *viper.Viper, mc *minio.Client, srcbucket, srcobject, dstbucket, dstobject string) error {

	// Use-case 1: Simple copy object with no conditions.
	// Source object
	srcOpts := minio.CopySrcOptions{
		Bucket: srcbucket,
		Object: srcobject,
	}

	// Destination object
	dstOpts := minio.CopyDestOptions{
		Bucket: dstbucket,
		Object: dstobject,
	}

	// Copy object call
	uploadInfo, err := mc.CopyObject(context.Background(), dstOpts, srcOpts)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Successfully copied object:", uploadInfo)

	return nil
}
