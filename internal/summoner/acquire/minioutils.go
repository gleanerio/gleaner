package acquire

import (
	"fmt"
	"log"

	"earthcube.org/Project418/gleaner/pkg/summoner/sitemaps"
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
				log.Printf("We already own %s, deleting current objects\n", bucketName)
			} else {
				log.Fatalln(err)
			}

			// TODO   should I empty the bucket if it exists?  (make this a flag?)
			objectsCh := make(chan string)
			// Send object names that are needed to be removed to objectsCh
			go func() {
				defer close(objectsCh)
				// List all objects from a bucket-name with a matching prefix.
				for object := range minioClient.ListObjects(bucketName, "", true, nil) {
					if object.Err != nil {
						log.Fatalln(object.Err)
					}
					objectsCh <- object.Key
				}
			}()

			for rErr := range minioClient.RemoveObjects(bucketName, objectsCh) {
				fmt.Println("Error detected during deletion: ", rErr)
			}
		}
		log.Printf("Successfully created %s\n", bucketName)

	}
}
