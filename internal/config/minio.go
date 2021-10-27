package config

import (
	"fmt"
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
	minioSubtress.AutomaticEnv()
	// config already read. substree passed
	err := minioSubtress.Unmarshal(&minioCfg)
	if err != nil {
		panic(fmt.Errorf("error when parsing minio config: %v", err))
	}
	return minioCfg, err
}
