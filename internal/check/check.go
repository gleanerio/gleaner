package check

import (
	"context"
	"fmt"
	"log"

	"github.com/minio/minio-go/v7"
)

var bl []string

// TODO  need to move this to the config file
func init() {
	bl = []string{"gleaner"}
	// bl := []string{"gleaner", "gleaner-config", "gleaner-summoned", "gleaner-milled", "gleaner-shacl", "gleaner-voc"}
	// bl = []string{"gleaner", "gleaner-summoned", "gleaner-milled", "gleaner-assets"}
}

// ConnCheck check the connections with a list buckets call
func ConnCheck(mc *minio.Client) error {
	_, err := mc.ListBuckets(context.Background())
	return err
}

// Buckets checks the setup
func Buckets(mc *minio.Client) error {
	var err error

	for i := range bl {
		found, err := mc.BucketExists(context.Background(), bl[i])
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("unable to locate required bucket:  %s, did you run gleaner with -setup the first to set up buckets?", bl[i])
		}
		if found {
			log.Printf("Validated access to object store: %s.\n", bl[i])
		}
	}

	return err
}

// MakeBuckets checks the setup
func MakeBuckets(mc *minio.Client) error {
	var err error

	for i := range bl {
		found, err := mc.BucketExists(context.Background(), bl[i])
		if err != nil {
			log.Printf("Existing bucket %s check:%v\n", bl[i], err)
		}
		if found {
			log.Printf("Gleaner Bucket %s found.\n", bl[i])
		} else {
			log.Printf("Gleaner Bucket %s not found, generating\n", bl[i])
			err = mc.MakeBucket(context.Background(), bl[i], minio.MakeBucketOptions{Region: "us-east-1"}) // location is kinda meaningless here
			if err != nil {
				log.Printf("Make bucket:%v\n", err)
			}
		}
	}

	return err
}

// // need to check, tika, shacl, headless
// // this is just a place holder for that work now
// // TODO this is just a placeholder not function..  needs to be finished
// // to loop on the services I need in place for gleaner
// func urlCheck() bool {
// 	s := false

// 	resp, err := http.Get("http://localhost:7001") // 9998  7000 32772
// 	if err != nil {
// 		return s
// 	}

// 	if resp.Status == "200 OK" {
// 		s = true
// 	}
// 	return s
// }

// // syscheck is a place holder function for work to be done....
// func syscheck(mc *minio.Client) {
// 	log.Println("System setup check placeholder")
// 	s := "NA"
// 	log.Printf("System check results: %s\n", s)
// }
