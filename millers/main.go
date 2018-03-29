package main

import (
	"fmt"
	"log"
	"time"

	"earthcube.org/Project418/gleaner/millers/millersgraph"
	"github.com/minio/minio-go"
)

func main() {
	fmt.Println("The miller....")
	st := time.Now()
	log.Printf("Start time: %s \n", st) // Log the time at start for the record

	mc := miniConnection() // minio connection

	// buckets, err := listBuckets(mc)
	// if err != nil {
	// 	log.Println(err)
	// }

	// fmt.Println("Bucket list...")
	// for _, bucket := range buckets {
	// 	fmt.Println(bucket.Name)
	// }

	// ----- MOCK call  (template )
	// millersmock.MockObjects(mc, "getiedadataorg")

	// ----- GRAPH calls (stores to file for each domain)
	millersgraph.GraphMillObjects(mc, "baltoopendaporg")
	millersgraph.GraphMillObjects(mc, "dataneotomadborg")
	millersgraph.GraphMillObjects(mc, "earthreforg")
	millersgraph.GraphMillObjects(mc, "getiedadataorg")
	millersgraph.GraphMillObjects(mc, "opencoredataorg")
	millersgraph.GraphMillObjects(mc, "opentopographyorg")
	millersgraph.GraphMillObjects(mc, "wikilinkedearth")
	millersgraph.GraphMillObjects(mc, "wwwbco-dmoorg")
	millersgraph.GraphMillObjects(mc, "wwwunavcoorg")

	// ----- SPATIAL calls (stores to tile38)
	// millerspatial.ProcessBucketObjects(mc, "opentopographyorg")
	// millerspatial.ProcessBucketObjects(mc, "dataneotomadborg")
	// millerspatial.ProcessBucketObjects(mc, "getiedadataorg")
	// millerspatial.ProcessBucketObjects(mc, "opencoredataorg")
	// millerspatial.ProcessBucketObjects(mc, "wwwbco-dmoorg")
	// millerspatial.ProcessBucketObjects(mc, "wikilinkedearth")

	// ----- ORGANIC index calls
	// millerbleve.GetObjects(mc, "opentopographyorg")
	// millerbleve.GetObjects(mc, "dataneotomadborg")
	// millerbleve.GetObjects(mc, "getiedadataorg")
	// millerbleve.GetObjects(mc, "opencoredataorg")
	// millerbleve.GetObjects(mc, "wwwbco-dmoorg")
	// millerbleve.GetObjects(mc, "wikilinkedearth")

	et := time.Now()
	diff := et.Sub(st)
	log.Printf("End time: %s \n", et)
	log.Printf("Run time: %f \n", diff.Minutes())
}

func miniConnection() *minio.Client {
	// Set up minio and initialize client
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
		fmt.Println(err)
		return nil, err
	}

	return buckets, err
}
