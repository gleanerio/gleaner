package config

import (
	"github.com/spf13/viper"
)

var gleanerTemplate = map[string]interface{}{
	"minio": map[string]string{
		"address":   "localhost",
		"port":      "9000",
		"accesskey": "",
		"secretkey": "",
	},
	"gleaner":     map[string]string{},
	"context":     map[string]string{},
	"contextmaps": map[string]string{},
	"summoner":    map[string]string{},
	"millers":     map[string]string{},
	"sources":     map[string]string{},
}

func ReadGleanerConfig(filename string, cfgDir string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range gleanerTemplate {
		v.SetDefault(key, value)
	}

	v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	v.AddConfigPath(cfgDir)
	v.SetConfigType("yaml")
	//v.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		panic(err)
	}
	return v, err
}
