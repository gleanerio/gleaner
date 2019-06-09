package check

import (
	"fmt"
	"log"
	"net/http"

	"github.com/minio/minio-go"
)

// GleanerSetup set up  elements of Gleaner are running

// GleanerCheck checks the setup
func GleanerCheck(mc *minio.Client) (bool, error) {
	var err error
	syscheck(mc)

	bl := []string{"gleaner", "gleaner-config", "gleaner-milled", "gleaner-shacl", "gleaner-voc"}

	for i := range bl {
		found, err := mc.BucketExists(bl[i])
		if err != nil {
			log.Printf("Existing bucket %s check:%v\n", bl[i], err)
		}

		if found {
			log.Printf("Gleaner Bucket %s found.\n", bl[i])
		} else {
			log.Printf("Gleaner Bucket %s not found, generating\n", bl[i])
			err = mc.MakeBucket(bl[i], "us-east-1") // location is kinda meaningless here
			if err != nil {
				log.Printf("Make bucket:%v\n", err)
			}
		}
	}

	// look for the docker containers like tika, shacl
	fmt.Printf("Checking for needed services in docker.\n (Not finished..  ignore results): %t \n", urlCheck())

	return true, err
}

// need to check, tika, shacl, headless
// this is just a place holder for that work now
// TODO this is just a placeholder not function..  needs to be finished
// to loop on the services I need in place for gleaner
func urlCheck() bool {
	s := false

	resp, err := http.Get("http://localhost:7001") // 9998  7000 32772
	if err != nil {
		return s
	}

	if resp.Status == "200 OK" {
		s = true
	}
	return s
}

// syscheck is a place holder function for work to be done....
func syscheck(mc *minio.Client) {
	fmt.Println("System setup check placeholder")
	s := "NA"
	fmt.Printf("System check results: %s\n", s)
}
