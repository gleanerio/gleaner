package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"io"
	"net/http"
	"strings"
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
		fmt.Printf("cannot find config file. '%v' If glcon Did you 'glcon generate --cfgName XXX' \n", filename)
		log.Fatalf("cannot find config file. '%v' Did you 'glcon generate --cfgName XXX' ", filename)
		//panic(err)
	}
	return v, err
}

//func ReadGleanerConfigURLAddRemote(configURL string) (*viper.Viper, error) {
//	v := viper.New()
//	for key, value := range gleanerTemplate {
//		v.SetDefault(key, value)
//	}
//
//	// Add the remote provider
//	err := v.AddRemoteProvider("https", configURL, "")
//	if err != nil {
//		log.Fatalf("Error adding remote provider: %v", err)
//	}
//
//	v.SetConfigType("yaml")
//
//	// Read the remote config
//	err = v.ReadRemoteConfig()
//	if err != nil {
//		log.Fatalf("Error reading remote config: %v", err)
//	}
//
//	fmt.Println("Config value:", v.Get("minio.address"))
//
//	//v.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
//	v.BindEnv("minio.address", "MINIO_ADDRESS")
//	v.BindEnv("minio.port", "MINIO_PORT")
//	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
//	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
//	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
//	//v.AutomaticEnv()
//	//err = v.ReadInConfig()
//	if err != nil {
//		fmt.Printf("cannot find config file. '%v' If glcon Did you 'glcon generate --cfgName XXX' \n", configURL)
//		log.Fatalf("cannot find config file. '%v' Did you 'glcon generate --cfgName XXX' ", configURL)
//		//panic(err)
//	}
//	return v, err
//}

func ReadGleanerConfigURL(configURL string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range gleanerTemplate {
		v.SetDefault(key, value)
	}

	resp, err := http.Get(configURL)
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return v, fmt.Errorf("HTTP request failed with status code %v", resp.StatusCode)
	}

	// Read the content of the config file
	configData, err := io.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}

	// Convert configData to a string
	configString := string(configData)

	// Convert the string to an io.Reader
	reader := strings.NewReader(configString)

	v.SetConfigType("yaml")
	err = v.ReadConfig(reader)
	if err != nil {
		fmt.Printf("Error reading config from URL: %v\n", err)
		return v, err
	}

	//value := v.GetString("minio.address")
	//fmt.Printf("Value of someKey: %v\n", value)

	//v.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	//v.AutomaticEnv()
	//err = v.ReadInConfig()
	if err != nil {
		fmt.Printf("cannot find config file. '%v' If glcon Did you 'glcon generate --cfgName XXX' \n", configURL)
		log.Fatalf("cannot find config file. '%v' Did you 'glcon generate --cfgName XXX' ", configURL)
		//panic(err)
	}
	return v, err
}
