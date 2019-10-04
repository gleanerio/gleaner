package common

import (
	"fmt"
	"log"

	minio "github.com/minio/minio-go"
	"github.com/spf13/viper"
)

// MinioConnection Set up minio and initialize client
func MinioConnection(v1 *viper.Viper) *minio.Client {
	mcfg := v1.GetStringMapString("minio")

	endpoint := fmt.Sprintf("%s:%s", mcfg["address"], mcfg["port"])
	accessKeyID := mcfg["accesskey"]
	secretAccessKey := mcfg["secretkey"]
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}
	return minioClient
}
