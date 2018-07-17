package utils

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

// Config struct
type Config struct {
	Minio struct {
		Endpoint        string `json:"endpoint"`
		AccessKeyID     string `json:"accessKeyID"`
		SecretAccessKey string `json:"secretAccessKey"`
	} `json:"minio"`
	Millers struct {
		Mock     bool `json:"mock"`
		Graph    bool `json:"graph"`
		Spatial  bool `json:"spatial"`
		Organic  bool `json:"organic"`
		Tika     bool `json:"tika"`
		FDPTika  bool `json:"fdptika"`
		FDPGraph bool `json:"fdpgraph"`
		Prov     bool `json:"prov"`
	} `json:"millers"`
	Sources []struct {
		Name          string `json:"name"`
		ShortName     string `json:"shortname"`
		URL           string `json:"url"`
		Headless      bool   `json:"headless"`
		Sitemapformat string `json:"sitemapformat"`
		Active        bool   `json:"active"`
	} `json:"sources"`
}

func LoadConfiguration(file *string) Config {
	var config Config
	configFile, err := os.Open(*file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

// DomainNameShort takes a URL, pulls them domain and removes the dots
func DomainNameShort(dn string) (string, string, error) {
	u, err := url.Parse(dn)
	if err != nil {
		panic(err)
	}

	// do we need to deal with port numbers from the host?
	// host, port, err := net.SplitHostPort(u.Host)
	// if err != nil {
	// 	log.Printf("Error parsing the domain name %v", err)
	// }

	// return host, port, u.Scheme, err

	// rewrite the host

	return strings.Replace(u.Host, ".", "", -1), u.Scheme, err
}

// GetSHA1 returns the sha1 string for the given byte array
func GetSHA1(b []byte) string {
	h := sha1.New()
	h.Write(b)
	bs := h.Sum(nil)
	return string(bs)
}
