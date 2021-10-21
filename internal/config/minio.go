package config

type Minio struct {
	address      string `mapstructure:"MINIO_ADDRESS"`
	port      string `mapstructure:"MINIO_PORT"`
	ssl string `mapstructure:"MINIO_USE_SSL"`
	accesskey string `mapstructure:"MINIO_ACCESS_KEY"`
	secretkey string `mapstructure:"MINIO_SECRET_KEY"`
	bucket string `mapstructure:"S3_BUCKET"`
}
