package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// frig frig... do not use lowercase... those are private variables
type Minio struct {
	Address   string // `mapstructure:"MINIO_ADDRESS"`
	Port      int    //`mapstructure:"MINIO_PORT"`
	Ssl       bool   //`mapstructure:"MINIO_USE_SSL"`
	Accesskey string //`mapstructure:"MINIO_ACCESS_KEY"`
	Secretkey string // `mapstructure:"MINIO_SECRET_KEY"`
	Bucket    string
}

var MinioTemplate = map[string]interface{}{
	"minio": map[string]string{
		"address":   "localhost",
		"port":      "9000",
		"accesskey": "",
		"secretkey": "",
		"bucket":    "",
	},
}

// use config.Sub("minio)
func ReadMinioConfig(minioSubtress *viper.Viper) (Minio, error) {
	var minioCfg Minio
	for key, value := range MinioTemplate {
		minioSubtress.SetDefault(key, value)
	}
	minioSubtress.BindEnv("address", "MINIO_ADDRESS")
	minioSubtress.BindEnv("port", "MINIO_PORT")
	minioSubtress.BindEnv("ssl", "MINIO_USE_SSL")
	minioSubtress.BindEnv("accesskey", "MINIO_ACCESS_KEY")
	minioSubtress.BindEnv("secretkey", "MINIO_SECRET_KEY")
	minioSubtress.BindEnv("secretkey", "MINIO_SECRET_KEY")
	minioSubtress.BindEnv("bucket", "MINIO_BUCKET")
	minioSubtress.AutomaticEnv()
	// config already read. substree passed
	err := minioSubtress.Unmarshal(&minioCfg)
	if err != nil {
		log.Fatal("error when parsing minio config: %v", err)
	}
	return minioCfg, err
}

func GetBucketName(v1 *viper.Viper) (string, error) {
	minSubtree := v1.Sub("minio")
	miniocfg, err := ReadMinioConfig(minSubtree)
	if err != nil {
		log.Fatal("Cannot read bucket name from configuration/minio")

	}
	bucketName := miniocfg.Bucket //miniocfg["bucket"] //   get the top level bucket for all of gleaner operations from config file
	return bucketName, err
}
