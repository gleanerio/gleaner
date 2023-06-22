package common

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	configTypes "github.com/gleanerio/gleaner/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

// MinioConnection Set up minio and initialize client
func MinioConnection(v1 *viper.Viper) *minio.Client {
	mSub := v1.Sub("minio")
	mcfg, err := configTypes.ReadMinioConfig(mSub)
	if err != nil {
		log.Panic("error when file minio key:", err)
	}

	var endpoint, accessKeyID, secretAccessKey string
	var useSSL bool

	if mcfg.Port == 0 {
		endpoint = fmt.Sprintf("%s", mcfg.Address)
		accessKeyID = mcfg.Accesskey
		secretAccessKey = mcfg.Secretkey
		useSSL = mcfg.Ssl
	} else {
		endpoint = fmt.Sprintf("%s:%d", mcfg.Address, mcfg.Port)
		accessKeyID = mcfg.Accesskey
		secretAccessKey = mcfg.Secretkey
		useSSL = mcfg.Ssl
	}

	var minioClient *minio.Client

	// used this == "" trick to not set region if not present due to
	// past issue of auth fails if a region is set in minioclient...
	if mcfg.Region == "" {
		log.Println("no region set")
		minioClient, err = minio.New(endpoint,
			&minio.Options{Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
				Secure: useSSL,
			})
	} else {
		log.Println("region set for GCS or AWS, may cause issues with minio")
		region := mcfg.Region
		minioClient, err = minio.New(endpoint,
			&minio.Options{Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
				Secure: useSSL,
				Region: region,
			})
	}

	// minioClient.SetCustomTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})
	if err != nil {
		log.Fatal(err)
	}

	return minioClient
}
