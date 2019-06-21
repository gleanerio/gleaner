package utils

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/minio/minio-go"
	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Gleaner struct {
		Summon     bool   `json:"summon"`
		Mill       bool   `json:"mill"`
		Configfile string `json:"configfile"`
		RunID      string `json:"runid"`
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
		NER         bool `json:"ner"`
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

// ConfigYAML take a string name of a configuration file
func ConfigYAML(file string) Config {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err.Error())
	}

	config := Config{}
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", config)

	return config
}

// S3ConfigYAML loads config file from S3(minio)
// func S3ConfigYAML(endpoint, port, accessKeyID, secretAccessKey, bucket, file string, useSSL bool) Config {
func S3ConfigYAML(minioClient *minio.Client, bucket, file string) Config {
	fo, err := minioClient.GetObject(bucket, file, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(fo)

	config := Config{}
	err = yaml.Unmarshal(buf.Bytes(), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- t:\n%v\n\n", config)

	return config
}

// LoadConfiguration take a string name of a configuration file
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
// func LoadConfigurationS3(endpoint, port, accessKeyID, secretAccessKey, bucket, file string, useSSL bool) Config {
func LoadConfigurationS3(minioClient *minio.Client, bucket, file string) Config {
	fo, err := minioClient.GetObject(bucket, file, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
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

	return strings.Replace(u.Host, ".", "", -1), u.Scheme, err
}

// GetSHA1 returns the sha1 string for the given byte array
func GetSHA1(b []byte) string {
	h := sha1.New()
	h.Write(b)
	bs := h.Sum(nil)
	return string(bs)
}

// MinioConnection Set up minio and initialize client
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

// ListBuckets list bucket at a minio server
func ListBuckets(mc *minio.Client) ([]minio.BucketInfo, error) {
	buckets, err := mc.ListBuckets()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return buckets, err
}
