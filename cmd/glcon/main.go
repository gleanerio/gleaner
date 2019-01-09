package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go"
)

var minioVal, portVal, accessVal, secretVal, bucketVal, cfgVal string
var sslVal bool
var check, cfginit, cfgval, cfgload, cfgrun bool

func init() {
	akey := os.Getenv("MINIO_ACCESS_KEY")
	skey := os.Getenv("MINIO_SECRET_KEY")

	flag.StringVar(&minioVal, "address", "localhost", "FQDN for server")
	flag.StringVar(&portVal, "port", "9000", "Port for minio server, default 9000")
	flag.StringVar(&accessVal, "access", akey, "Access Key ID")
	flag.StringVar(&secretVal, "secret", skey, "Secret access key")
	flag.StringVar(&bucketVal, "bucket", "gleaner", "The configuration bucket")
	flag.StringVar(&cfgVal, "config", "config.json", "Configuration file")
	flag.BoolVar(&sslVal, "ssl", false, "Use SSL boolean")

	flag.BoolVar(&check, "check", false, "Check the running minio setup")
	flag.BoolVar(&cfginit, "cfginit", false, "Init a config template")
	flag.BoolVar(&cfgval, "cfgval", false, "Validate a config file is parsable")

	// cfgload is NOT boolean, it's a string pointing to the config file to load and run
	flag.BoolVar(&cfgload, "cfgload", false, "Load the config into the mino input queue bucket")

	// cfgrun is NOT needed..  load will validate and load..  the buckets will trigger
	// the webhook call to gleaner to check the input queue
	flag.BoolVar(&cfgrun, "cfgrun", false, "Call Gleaner to check input queue bucket")
}

func main() {
	fmt.Println("EarthCube Gleaner Control")
	flag.Parse()

	// init s3 client
	// init s3 buckets
	err := initBucket()
	if err != nil {
		fmt.Println(err)
	}

	// init config
	if cfginit {
		err := initCfg()
		if err != nil {
			fmt.Println(err)
		}
	}

	// validate config
	if cfgval {
		err := valCfg()
		if err != nil {
			fmt.Println(err)
		}
	}

	// load config, on load minio will call webhook in gleaner to run config
	if cfgload {
		err := loadCfg()
		if err != nil {
			fmt.Println(err)
		}
	}

}

func initCfg() error {
	fmt.Println("make a config template is there isn't already one")
	return nil
}

func valCfg() error {
	fmt.Println("validate config")
	return nil
}

func loadCfg() error {
	fmt.Println("load config")
	return nil
}

func initBucket() error {

	// Set up minio and initialize client
	endpoint := "192.168.2.131:9000"
	accessKeyID := "AKIAIOSFODNN7EXAMPLE"
	secretAccessKey := "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	useSSL := false
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	// Make a new bucket called mymusic.
	bucketName := "onlyatest"
	location := ""

	err = minioClient.MakeBucket(bucketName, location)
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, err := minioClient.BucketExists(bucketName)
		if err == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	err = setNotification(minioClient)
	if err != nil {
		log.Fatalln(err)
	}

	return err
}

func setNotification(minioClient *minio.Client) error {
	queueArn := minio.NewArn("minio", "sqs", "", "1", "webhook")
	// arn:minio:sqs::1:webhook

	queueConfig := minio.NewNotificationConfig(queueArn)
	queueConfig.AddEvents(minio.ObjectCreatedAll, minio.ObjectRemovedAll)
	// queueConfig.AddFilterPrefix("photos/")
	queueConfig.AddFilterSuffix(".json")

	bucketNotification := minio.BucketNotification{}
	bucketNotification.AddQueue(queueConfig)

	err := minioClient.SetBucketNotification("onlyatest", bucketNotification)
	if err != nil {
		fmt.Println("Unable to set the bucket notification: ", err)
		return nil
	}

	return nil
}
