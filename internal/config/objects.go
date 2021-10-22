package config

type Objects struct {
	bucket    string `mapstructure:"S3_BUCKET"`
	domain    string `mapstructure:"S3_DOMAIN"`
	prefix    []string
	prefixOff []string
}
