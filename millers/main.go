package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"earthcube.org/Project418/gleaner/millers/millerbleve"
	"earthcube.org/Project418/gleaner/millers/millersgraph"
	"earthcube.org/Project418/gleaner/millers/millersmock"
	"earthcube.org/Project418/gleaner/millers/millerspatial"
	"earthcube.org/Project418/gleaner/millers/utils"

	"github.com/minio/minio-go"
)

func main() {
	// Set up our log file for runs...
	f, err := os.OpenFile("logfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Println("The miller....")
	st := time.Now()
	log.Printf("Start time: %s \n", st) // Log the time at start for the record

	mc := miniConnection() // minio connection

	// Print out the buckets available to work with (for no real reason....)
	buckets, err := listBuckets(mc)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("Bucket list...")
	for _, bucket := range buckets {
		log.Println(bucket.Name)
	}

	// load configuration
	cfgFileLoc := flag.String("config", "config.json", "JSON configure file")
	flag.Parse()

	cs := loadConfiguration(cfgFileLoc)
	as := []string{}

	for i := range cs.Sources {
		if cs.Sources[i].Active == true {
			as = append(as, cs.Sources[i].Name)
		}
	}

	if cs.Millers.Mock {
		for d := range as {
			millersmock.MockObjects(mc, as[d])
		}
	}

	if cs.Millers.Graph {
		for d := range as {
			millersgraph.GraphMillObjects(mc, as[d])
		}
	}

	if cs.Millers.Spatial {
		for d := range as {
			millerspatial.ProcessBucketObjects(mc, as[d])
		}
	}

	if cs.Millers.Organic {
		for d := range as {
			millerbleve.GetObjects(mc, as[d])
		}
	}

	et := time.Now()
	diff := et.Sub(st)
	log.Printf("End time: %s \n", et)
	log.Printf("Run time: %f \n", diff.Minutes())
}

func loadConfiguration(file *string) utils.Config {
	var config utils.Config
	configFile, err := os.Open(*file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

// Set up minio and initialize client
func miniConnection() *minio.Client {
	endpoint := "localhost:9000"
	accessKeyID := "AKIAIOSFODNN7EXAMPLE"
	secretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	return minioClient
}

func listBuckets(mc *minio.Client) ([]minio.BucketInfo, error) {
	buckets, err := mc.ListBuckets()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return buckets, err
}
