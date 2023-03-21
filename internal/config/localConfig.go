package config

import (
	"github.com/spf13/viper"
)

// auth fails if a region is set in minioclient...
var serversTemplate = map[string]interface{}{
	"minio": map[string]string{
		"address":   "localhost",
		"port":      "9000",
		"accesskey": "",
		"secretkey": "",
		"bucket":    "",
		//		"region":    "us-east-1",
	},
	"sparql": map[string]string{
		"endpoint": "localhost",
	},
	"headless": "",
	"s3": map[string]string{
		"bucket": "gleaner",
		"domain": "us-east-1",
	},
	"identifiertype": JsonSha, // const from config.Sources jsonsha,identifiersha, normalizedjsonsha, identifierstring
}

func ReadServersConfig(filename string, cfgDir string) (*viper.Viper, error) {
	v := viper.New()
	for key, value := range serversTemplate {
		v.SetDefault(key, value)
	}
	v.BindEnv("minio.address", "MINIO_ADDRESS")
	v.BindEnv("minio.port", "MINIO_PORT")
	v.BindEnv("minio.ssl", "MINIO_USE_SSL")
	v.BindEnv("minio.accesskey", "MINIO_ACCESS_KEY")
	v.BindEnv("minio.secretkey", "MINIO_SECRET_KEY")
	v.BindEnv("minio.bucket", "MINIO_BUCKET")
	//	v.BindEnv("minio.region", "MINIO_REGION")
	v.BindEnv("sparql.endpoint", "SPARQL_ENDPOINT")
	v.BindEnv("sparql.authenticate", "SPARQL_AUTHENTICATE")
	v.BindEnv("sparql.username", "SPARQL_USERNAME")
	v.BindEnv("sparql.password", "SPARQL_PASSWORD")
	v.BindEnv("s3.domain", "S3_DOMAIN")

	v.SetConfigName(fileNameWithoutExtTrimSuffix(filename))
	v.AddConfigPath(cfgDir)
	v.SetConfigType("yaml")
	//v.BindEnv("headless", "GLEANER_HEADLESS_ENDPOINT")
	v.AutomaticEnv()
	err := v.ReadInConfig()
	return v, err
}
