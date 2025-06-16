package config

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"strings"
)

var nabuTemplate = map[string]interface{}{
	"minio":   MinioTemplate,
	"sparql":  sparqlTemplate,
	"objects": ObjectTemplate,
}

func ReadNabuConfig(filename string, cfgPath string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range nabuTemplate {
		v.SetDefault(key, value)
	}

	v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	v.AddConfigPath(cfgPath)
	v.SetConfigType("yaml")
	//v.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	v.BindEnv("minio.bucket", "MINIO_BUCKET")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}

func ReadNabuConfigURL(configURL string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range nabuTemplate {
		v.SetDefault(key, value)
	}

	log.Printf("Reading config from URL: %v\n", configURL)

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

	//v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	//v.AddConfigPath(cfgPath)
	v.SetConfigType("yaml")
	//v.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	v.BindEnv("minio.bucket", "MINIO_BUCKET")
	v.AutomaticEnv()

	err = v.ReadConfig(reader)
	if err != nil {
		fmt.Printf("Error reading config from URL: %v\n", err)
		return v, err
	}

	return v, err
}
