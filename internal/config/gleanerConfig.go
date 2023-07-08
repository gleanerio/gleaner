package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path"
)

// auth fails if a region is set in minioclient...
var gleanerTemplate = map[string]interface{}{
	"minio": map[string]string{
		"address":   "localhost",
		"port":      "9000",
		"region":    "",
		"accesskey": "",
		"secretkey": "",
	},
	"gleaner":     map[string]string{},
	"context":     map[string]string{},
	"contextmaps": map[string]string{},
	"summoner":    map[string]string{},
	"millers":     map[string]string{},
	"sources": map[string]string{
		"sourcetype": "sitemap",
		"name":       "",
		"url":        "",
		"logo":       "",
		"headless":   "",
		"pid":        "",
		"propername": "",
		"domain":     "",
	},
}

func ReadGleanerConfig(filename string, cfgDir string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range gleanerTemplate {
		v.SetDefault(key, value)
	}
	if _, err := os.Stat(path.Join(cfgDir, filename)); err != nil {
		fmt.Printf("File does not exist\n")
		fmt.Printf("cannot find config file. '%v' If glcon Did you 'glcon generate --cfgName XXX' \n", filename)
		log.Fatalf("cannot find config file. '%v' Did you 'glcon generate --cfgName XXX' ", filename)
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
		fmt.Printf("Error Reading Config: {err}")
		log.Fatalf("Error Reading Config: {err}")
		//panic(err)
	}
	return v, err
}
