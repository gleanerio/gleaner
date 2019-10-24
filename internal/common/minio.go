package common

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	minio "github.com/minio/minio-go"
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
	minioClient, err := minio.NewV2(endpoint, accessKeyID, secretAccessKey, useSSL)
	minioClient.SetCustomTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})
	if err != nil {
		log.Fatalln(err)
	}
	return minioClient
}
