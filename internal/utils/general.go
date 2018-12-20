package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go"
)

// Config struct
type Config struct {
	Gleaner struct {
		Summon     bool   `json:"summon"`
		Mill       bool   `json:"mill"`
		Configfile string `json:"configfile"`
	} `json:"gleaner"`
	Minio struct {
		Endpoint        string `json:"endpoint"`
		AccessKeyID     string `json:"accessKeyID"`
		SecretAccessKey string `json:"secretAccessKey"`
	} `json:"minio"`
	Millers struct {
		Mock        bool `json:"mock"`
		Graph       bool `json:"graph"`
		Spatial     bool `json:"spatial"`
		Shacl       bool `json:"shacl"`
		Organic     bool `json:"organic"`
		Tika        bool `json:"tika"`
		FDPTika     bool `json:"fdptika"`
		FDPTikaJena bool `json:"fdptikajena"`
		FDPGraph    bool `json:"fdpgraph"`
		Prov        bool `json:"prov"`
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

//LoadConfiguration take a string name of a configuration file
func LoadConfiguration(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

// LoadConfigurationS3 loads config file from S3(minio)
func LoadConfigurationS3(endpoint, port, accessKeyID, secretAccessKey, bucket, file string, useSSL bool) Config {
	// // Set up minio and initialize client
	ep := fmt.Sprintf("%s:%s", endpoint, port)
	minioClient, err := minio.New(ep, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	// log.Println(minioClient.ListBuckets())

	fo, err := minioClient.GetObject(bucket, file, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(fo)

	var config Config
	jsonParser := json.NewDecoder(buf)
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

// Set up minio and initialize client
func MinioConnection(cs Config) *minio.Client {
	endpoint := cs.Minio.Endpoint
	accessKeyID := cs.Minio.AccessKeyID
	secretAccessKey := cs.Minio.SecretAccessKey
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}
	return minioClient
}

func ListBuckets(mc *minio.Client) ([]minio.BucketInfo, error) {
	buckets, err := mc.ListBuckets()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buckets, err
}
