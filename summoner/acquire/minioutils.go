package acquire

import (
	"log"

	"earthcube.org/Project418/gleaner/summoner/sitemaps"
	"github.com/minio/minio-go"
)

func buildBuckets(minioClient *minio.Client, m map[string]sitemaps.URLSet) {

	log.Println("Building buckets")
	for k := range m {

		// get the base domain from the map

		// minio buckets   make one for each domain in map
		// make a separate function in the assemble package
		bucketName := k
		location := "us-east-1"
		err := minioClient.MakeBucket(bucketName, location)
		if err != nil {
			// Check to see if we already own this bucket (which happens if you run this twice)
			exists, err := minioClient.BucketExists(bucketName)
			if err == nil && exists {
				log.Printf("We already own %s\n", bucketName)
			} else {
				log.Fatalln(err)
			}
		}
		log.Printf("Successfully created %s\n", bucketName)

	}
}
