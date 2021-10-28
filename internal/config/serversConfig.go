package config

import (
	"github.com/spf13/viper"
)

var serversTemplate = map[string]interface{}{
	"minio": map[string]string{
		"address":   "localhost",
		"port":      "9000",
		"accesskey": "",
		"secretkey": "",
		"bucket":    "",
	},
	"sparql": map[string]string{
		"endpoint": "localhost",
	},
	"headless": "",
	"s3": map[string]string{
		"bucket": "gleaner",
		"domain": "us-east-1",
	},
}

func ReadServersConfig(filename string, cfgDir string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range serversTemplate {
		v.SetDefault(key, value)
	}

	v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	v.AddConfigPath(cfgDir)
	v.SetConfigType("yaml")
	//v.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}
