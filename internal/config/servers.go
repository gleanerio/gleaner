package config

var ServersTemplate = map[string]interface{}{
	"minio": map[string]string{
		"address":   "localhost",
		"port":      "9000",
		"accesskey": "",
		"secretkey": "",
	},
	"sparql": map[string]string{
		"endpoint": "localhost",
	},
	"headless": "",
}
