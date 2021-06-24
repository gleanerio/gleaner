package common

import (
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

// MinioConnection Set up minio and initialize client
func MinioConnection(v1 *viper.Viper) *minio.Client {
	//mcfg := v1.GetStringMapString("minio")
	mcfg := v1.Sub("minio")
	endpoint := fmt.Sprintf("%s:%s", mcfg.GetString("address"), mcfg.GetString("port"))
	accessKeyID := mcfg.GetString("accesskey")
	secretAccessKey := mcfg.GetString("secretkey")
	useSSL := mcfg.GetBool("ssl")

	minioClient, err := minio.New(endpoint,
		&minio.Options{Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
			Secure: useSSL})
	// minioClient.SetCustomTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})
	if err != nil {
		log.Fatalln(err)
	}
	return minioClient
}
