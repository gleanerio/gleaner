package config

import "github.com/spf13/viper"

var nabuTemplate = map[string]interface{}{
	"minio":   "",
	"sparql":  "",
	"objects": "",
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
