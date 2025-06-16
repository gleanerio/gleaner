package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Objects struct {
	Bucket string //`mapstructure:"MINIO_BUCKET"`
	Domain string //`mapstructure:"MINIO_DOMAIN"`
	Prefix []string
}

var ObjectTemplate = map[string]interface{}{
	"objects": map[string]interface{}{
		"bucket": "gleaner",
		"domain": "us-east-1",
		"prefix": map[string][]string{},
	},
}

func GetObjectsConfig(viperConfig *viper.Viper) (Objects, error) {
	sub := viperConfig.Sub("objects")
	return ReadObjectsConfig(sub)
}

func ReadObjectsConfig(viperSubtree *viper.Viper) (Objects, error) {
	var objects Objects
	for key, value := range sparqlTemplate {
		viperSubtree.SetDefault(key, value)
	}
	viperSubtree.BindEnv("bucket", "MINIO_BUCKET")
	viperSubtree.BindEnv("domain", "S3_DOMAIN")
	viperSubtree.AutomaticEnv()
	// config already read. substree passed
	err := viperSubtree.Unmarshal(&objects)
	if err != nil {
		panic(fmt.Errorf("error when parsing servers s3  config: %v", err))
	}
	return objects, err
}
